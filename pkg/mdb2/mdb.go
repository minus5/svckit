package mdb2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"strings"
	"text/template"
	"time"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

var (
	// ErrNotFound represents error if document doesn't exist
	ErrNotFound = errors.New("not found")
	// ErrDuplicate represents error if document already exists
	ErrDuplicate = errors.New("duplicate document")
)

// Mdb is struct for handling mongo connection
type Mdb struct {
	client       *mongo.Client
	clientOptions *options.ClientOptions
	db           *mongo.Database
	checkPointIn time.Duration
	name         string
	cacheDir     string
	cache        *cache
}

// DefaultConnStr creates connection string from consul
func DefaultConnStr() string {
	// cita iz mongo kv store key mongo
	cs, err := dcy.KV("mongo/default/connectionString")
	if err == nil && cs != "" {
		app := env.AppName()
		kvs, err := dcy.KVs("mongo/" + app)
		_, disabled := kvs["disabled"]
		if err == nil && !disabled {
			return connectionStringFromTemplate(cs, kvs["database"], kvs["username"], kvs["password"])
		}
	}

	connStr := "mongo.service.sd"
	if addrs, err := dcy.LocalServices(connStr); err == nil {
		connStr = strings.Join(addrs.String(), ",")
	}
	return connStr
}

func connectionStringFromTemplate(tpl, database, username, password string) string {
	param := struct {
		Database string
		Username string
		Password string
	}{
		database,
		username,
		password,
	}

	buf := bytes.NewBuffer(nil)
	pt := template.Must(template.New("").Parse(tpl))
	if err := pt.Execute(buf, param); err != nil {
		log.Error(err)
		return ""
	}

	return buf.String()
}

// MustNew raises fatal is unable to connect to mongo
func MustNew(connStr string, opts ...func(*Mdb)) *Mdb {
	db, err := NewDb(connStr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Name sets name of the database
func Name(n string) func(*Mdb) {
	return func(mdb *Mdb) {
		mdb.name = n
	}
}

func SetPoolLimit(limit int) func(*Mdb) {
	return func(mdb *Mdb) {
		mdb.clientOptions.SetMaxPoolSize(uint64(limit))
	}
}

// CacheRoot sets disk cache root directory
func CacheRoot(d string) func(*Mdb) {
	return func(mdb *Mdb) {
		if d != "" {
			mdb.cacheDir = fmt.Sprintf("%s/%s", d, mdb.name)
		}
	}
}

// EnsureSafe sets session into Safe mode - ensure session is at least checking for errors, without enforcing further constraints
// request acknowledgment that write operations propagated to at least 1 mongod instance
// the closest I could find to replace EnsureSafe(&mgo.Safe{}) from mgo driver
func EnsureSafe() func(*Mdb) {
	return func(mdb *Mdb) {
		mdb.clientOptions.SetWriteConcern(writeconcern.New(writeconcern.W(1)))
	}
}

// MajoritySafe  requests acknowledgement that write operations propagated to the majority of mongod instances
func MajoritySafe() func(mdb *Mdb) {
	return func(mdb *Mdb) {
		mdb.clientOptions.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	}
}

// SetModePrimaryPreferred sets mode to primary preferred
func SetModePrimaryPreferred() func(mdb *Mdb) {
	return func(mdb *Mdb) {
		mdb.clientOptions.SetReadPreference(readpref.Primary())
	}
}

// CacheCheckpoint sets checkpoint interval
// when all cached items are flushed into mongo
func CacheCheckpoint(d time.Duration) func(mdb *Mdb) {
	return func(mdb *Mdb) {
		mdb.checkPointIn = d
	}
}

// NewDb creates new Db
func NewDb(connStr string, opts ...func(db *Mdb)) (*Mdb, error) {
	mdb := &Mdb{}
	if err := mdb.Init(connStr, opts...); err != nil {
		return nil, err
	}
	return mdb, nil
}

// Init initializes new Mdb
// Connects to mongo, initializes cache, starts checkpoint loop.
func (mdb *Mdb) Init(connStr string, opts ...func(db *Mdb)) error {
	mdb.checkpoint()
	mdb.clientOptions = options.Client().
		ApplyURI(connStr).
		SetReadPreference(readpref.SecondaryPreferred()).
		// don't wait for acknowledgment that write operations propagated to the any of the mongod instances
		// SetSafe(nil) from mgo driver
		SetWriteConcern(writeconcern.New(writeconcern.W(0)))
	mdb.checkPointIn = time.Minute
	mdb.name = strings.Replace(env.AppName(), ".", "_", -1)

	for _, opt := range opts {
		opt(mdb)
	}

	if err := mdb.clientOptions.Validate(); err != nil {
		return err
	}

	client, err := mongo.Connect(context.Background(), mdb.clientOptions)
	if err != nil {
		return err
	}

	if mdb.cacheDir != "" {
		mdb.cache, err = newCache(mdb)
		if err != nil {
			return err
		}
		go mdb.loop()
	}

	mdb.client = client
	mdb.db = client.Database(mdb.name)
	return nil
}

func (mdb *Mdb) loop() {
	t := time.NewTicker(mdb.checkPointIn)
	for range t.C {
		mdb.checkpoint()
	}
}

func (mdb *Mdb) checkpoint() {
	if mdb.cache != nil {
		mdb.cache.purge()
	}
}

// Ping checks if mongo is available
func (mdb *Mdb) Ping() bool {
	return mdb.client.Ping(context.Background(), nil) == nil
}

// Use wraps handler function with timing metric
func (mdb *Mdb) Use(col string, metricKey string, handler func(*mongo.Collection) error) error {
	c := mdb.db.Collection(col)
	var err error
	metric.Timing("db."+metricKey, func() {
		err = handler(c)
	})
	return err
}

// Use2 is same as Use but without metricKey
// metricKey is set to collection name
func (mdb *Mdb) Use2(col string, handler func(*mongo.Collection) error) error {
	return mdb.Use(col, col, handler)
}

// ReadId reads document with specified id from mongo
func (mdb *Mdb) ReadId(col string, id interface{}, o interface{}) error {
	if mdb.cache != nil {
		if i, ok := mdb.cache.find(col, id); ok {
			return i.unmarshal(o)
		}
	}
	return mdb.Use(col, "saveId", func(c *mongo.Collection) error {
		err := c.FindOne(context.Background(), bson.D{{"_id", id}}).Decode(o)
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return err
	})
}

// SaveId stores document to mongo
func (mdb *Mdb) SaveId(col string, id interface{}, o interface{}) error {
	if mdb.cache != nil {
		return mdb.cache.add(col, id, o)
	}
	return mdb.saveId(col, id, o)
}

func (mdb *Mdb) saveId(col string, id interface{}, o interface{}) error {
	return mdb.Use(col, "saveId", func(c *mongo.Collection) error {
		_, err := c.UpdateOne(context.Background(), bson.D{{"_id", id}}, o, options.Update().SetUpsert(true))
		return err
	})
}

// Exists checks if document matching specified query exists in mongo
func (mdb *Mdb) Exists(col string, query interface{}) (bool, error) {
	exists := false
	err := mdb.Use(col, "exists", func(c *mongo.Collection) error {
		count, err := c.CountDocuments(context.Background(), query)
		exists = count > 0
		return err
	})
	return exists, err
}

// RemoveId removes document with specified id from mongo
func (mdb *Mdb) RemoveId(col string, id interface{}) error {
	return mdb.Use(col, col+"remove", func(c *mongo.Collection) error {
		_, err := c.DeleteOne(context.Background(), bson.D{{"_id", id}})
		return err
	})
}

// Insert inserts new document to mongo
func (mdb *Mdb) Insert(col string, o interface{}) error {
	return mdb.Use(col, col+"insert", func(c *mongo.Collection) error {
		_, err := c.InsertOne(context.Background(), o)
		// TODO check duplicate
		return err
	})
}

// Close starts clean exit
func (mdb *Mdb) Close() {
	mdb.checkpoint()
}

func (mdb *Mdb) ResetIndexCache() {
	// TODO
}

// EnsureIndex creates index if it doesn't already exist
func (mdb *Mdb) EnsureIndex(col string, key []string, expireAfter time.Duration) error {
	c := mdb.db.Collection(col)
	parsedKeys, err := parseIndexKey(key)
	if err != nil {
		return err
	}
	index := mongo.IndexModel{}
	index.Keys = parsedKeys.key
	options := options.Index().
		SetBackground(true).
		SetSparse(true).
		SetName(parsedKeys.name)
	if expireAfter > 0 {
		options = options.SetExpireAfterSeconds(int32(expireAfter / time.Second))
	}
	index.Options = options
	_, err = c.Indexes().CreateOne(context.Background(), index)
	return err
}

// EnsureUniqueIndex creates unique index if it doesn't already exist
func (mdb *Mdb) EnsureUniqueIndex(col string, key []string) error {
	c := mdb.db.Collection(col)
	parsedKeys, err := parseIndexKey(key)
	if err != nil {
		return err
	}
	index := mongo.IndexModel{}
	index.Keys = parsedKeys.key
	index.Options = options.Index().
		SetBackground(true).
		SetSparse(true).
		SetName(parsedKeys.name).
		SetUnique(true)
	_, err = c.Indexes().CreateOne(context.Background(), index)
	return err
}

type indexKeyInfo struct {
	name string
	key  bsonx.Doc
}

// parseIndexKey provides backward compatibility with mgo driver for index definition
func parseIndexKey(key []string) (*indexKeyInfo, error) {
	var keyInfo indexKeyInfo
	isText := false
	var order int32
	for _, field := range key {
		raw := field
		if keyInfo.name != "" {
			keyInfo.name += "_"
		}
		var kind string
		if field != "" {
			if field[0] == '$' {
				if c := strings.Index(field, ":"); c > 1 && c < len(field)-1 {
					kind = field[1:c]
					field = field[c+1:]
					keyInfo.name += field + "_" + kind
				} else {
					field = "\x00"
				}
			}
			switch field[0] {
			case 0:
				// Logic above failed. Reset and error.
				field = ""
			case '-':
				order = -1
				field = field[1:]
				keyInfo.name += field + "_-1"
			case '+':
				field = field[1:]
				fallthrough
			default:
				if kind == "" {
					order = 1
					keyInfo.name += field + "_1"
				}
			}
		}
		if field == "" {
			return nil, fmt.Errorf(`invalid index key: want "[$<kind>:][-]<field name>", got %q`, raw)
		}
		if kind == "text" {
			if !isText {
				keyInfo.key.Append("_fts", bsonx.String("text"))
				keyInfo.key.Append("_ftsx", bsonx.Int32(1))
				isText = true
			}
		} else if kind != "" {
			keyInfo.key.Append(field, bsonx.String(kind))
		} else {
			keyInfo.key.Append(field, bsonx.Int32(order))
		}
	}
	if keyInfo.name == "" {
		return nil, errors.New("invalid index key: no fields provided")
	}
	return &keyInfo, nil
}
