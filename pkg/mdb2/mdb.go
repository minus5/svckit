package mdb2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/minus5/svckit/asm"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// ErrNotFound represents error if document doesn't exist
	ErrNotFound = errors.New("not found")
	// ErrDuplicate represents error if document already exists
	ErrDuplicate = errors.New("duplicate document")
)

// Mdb is struct for handling mongo connection
type Mdb struct {
	client        *mongo.Client
	clientOptions *options.ClientOptions
	db            *mongo.Database
	checkPointIn  time.Duration
	name          string
	cacheDir      string
	cache         *cache
}

// DefaultConnStr creates connection string from consul
func DefaultConnStr() string {
	// cita iz mongo kv store key mongo
	var cs string
	app := env.AppName()
	if acs, err := dcy.KV(fmt.Sprintf("mongo/%s/connectionString", app)); err == nil && acs != "" {
		cs = acs
		log.Info("using custom mongo connection string - %s", cs)
	}
	if cs == "" {
		if dcs, err := dcy.KV("mongo/default/connectionString"); err == nil && dcs != "" {
			cs = dcs
		}
	}
	if cs != "" {
		kvs, err := fetchKV("mongo/" + app)
		_, disabled := kvs["disabled"]
		if err == nil && !disabled {
			return connectionStringFromTemplate(cs, kvs["database"], kvs["username"], kvs["password"])
		}
	}

	connStr := "mongo.service.sd"
	if addrs, err := dcy.LocalServices(connStr); err == nil {
		connStr = fmt.Sprintf("mongodb://%s", strings.Join(addrs.String(), ","))
	}
	return connStr
}

func fetchKV(name string) (map[string]string, error) {
	kvs := map[string]string{}
	err := asm.ParseKV(name, &kvs)
	log.S("name", name).I("len", len(kvs)).Info("ASM fetched")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(kvs) > 0 {
		return kvs, nil
	}
	return dcy.KVs(name)
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

	// driver defaults to decoding interface{} as bson.D whereas mgo defaults to bson.M
	// this code matches mgo behavior and allows to decode interface{} as JSON
	// https://jira.mongodb.org/browse/GODRIVER-988
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).
		// Read int32 as int (when deserializing into interface{}) for backward compatibility
		RegisterTypeMapEntry(bsontype.Int32, reflect.TypeOf(int(0))).
		// Read bson.Array as []interface{} for backward compatibility
		RegisterTypeMapEntry(bsontype.Array, reflect.TypeOf([]interface{}{})).
		Build()
	mdb.clientOptions.SetRegistry(reg)

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

func (mdb *Mdb) UseSafe(col string, metricKey string, handler func(*mongo.Collection) error) error {
	// create new database object so options can be changed regardless of original
	c := mdb.client.Database(mdb.name, options.Database().SetWriteConcern(writeconcern.New(writeconcern.WMajority()))).Collection(col)
	var err error
	metric.Timing("db."+metricKey, func() {
		err = handler(c)
	})
	return err
}

func (mdb *Mdb) UseWithoutTimeout(col string, handler func(*mongo.Collection) error) error {
	// this option will probably have to go through options passed to Find and FindOne - SetNoCursorTimeout
	// eg. options.Find().SetNoCursorTimeout(true)
	return fmt.Errorf("not implemented")
}

// ReadId reads document with specified id from mongo
func (mdb *Mdb) ReadId(col string, id interface{}, o interface{}, metrics ...string) error {
	if mdb.cache != nil {
		if i, ok := mdb.cache.find(col, id); ok {
			return i.unmarshal(o)
		}
	}
	return mdb.Use(col, getMetricKey("readId", metrics...), func(c *mongo.Collection) error {
		sr := c.FindOne(context.Background(), bson.D{{"_id", id}})
		err := sr.Err()
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		return sr.Decode(o)
	})
}

// SaveId stores document to mongo
func (mdb *Mdb) SaveId(col string, id interface{}, o interface{}, metrics ...string) error {
	if mdb.cache != nil {
		return mdb.cache.add(col, id, o)
	}
	return mdb.saveId(col, getMetricKey("saveId", metrics...), id, o)
}

func (mdb *Mdb) saveId(col, metricKey string, id interface{}, o interface{}) error {
	return mdb.Use(col, metricKey, func(c *mongo.Collection) error {
		_, err := c.ReplaceOne(context.Background(), bson.D{{"_id", id}}, o, options.Replace().SetUpsert(true))
		return err
	})
}

// Exists checks if document matching specified query exists in mongo
func (mdb *Mdb) Exists(col string, query interface{}, metrics ...string) (bool, error) {
	exists := false
	err := mdb.Use(col, getMetricKey("exists", metrics...), func(c *mongo.Collection) error {
		count, err := c.CountDocuments(context.Background(), query)
		exists = count > 0
		return err
	})
	return exists, err
}

// RemoveId removes document with specified id from mongo
func (mdb *Mdb) RemoveId(col string, id interface{}, metrics ...string) error {
	if mdb.cache != nil {
		mdb.cache.remove(col, id)
	}
	return mdb.Use(col, getMetricKey(col+"remove", metrics...), func(c *mongo.Collection) error {
		dr, err := c.DeleteOne(context.Background(), bson.D{{"_id", id}})
		if err != nil {
			return err
		}
		if dr.DeletedCount == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// Insert inserts new document to mongo
func (mdb *Mdb) Insert(col string, o interface{}, metrics ...string) error {
	return mdb.Use(col, getMetricKey(col+"insert", metrics...), func(c *mongo.Collection) error {
		_, err := c.InsertOne(context.Background(), o)
		if IsDup(err) {
			return ErrDuplicate
		}
		return err
	})
}

func (mdb *Mdb) DropCollection(col string) error {
	return mdb.Use(col, col+"drop", func(c *mongo.Collection) error {
		return c.Drop(context.Background())
	})
}

// Close starts clean exit
func (mdb *Mdb) Close() {
	mdb.checkpoint()
}

func (mdb *Mdb) ResetIndexCache() {
	// probably not needed but leaving for backward compatibility for now
}

// EnsureIndex creates sparse index if it doesn't already exist
// The index will by default be sparse and built in background
func (mdb *Mdb) EnsureIndex(col string, key []string, expireAfter time.Duration) error {
	opts := options.Index().
		SetBackground(true).
		SetSparse(true)
	if expireAfter > 0 {
		opts.SetExpireAfterSeconds(int32(expireAfter / time.Second))
	}

	return mdb.ensureIndex(col, key, opts)
}

// EnsureCustomIndex creates index with the specified options if it doesn't already exist
// NOTE: IndexOptions#Name will always be overridden to obey legacy naming system derived from key values
func (mdb *Mdb) EnsureCustomIndex(col string, key []string, options *options.IndexOptions) error {
	return mdb.ensureIndex(col, key, options)
}

// ensureIndex ensures index exist. Uses legacy key parsing for backward compatibility meaning that
// name is automatically set from the key values
func (mdb *Mdb) ensureIndex(col string, key []string, indexOptions *options.IndexOptions) error {
	c := mdb.db.Collection(col)
	parsedKeys, err := parseIndexKey(key)
	if err != nil {
		return err
	}
	index := mongo.IndexModel{}
	index.Keys = parsedKeys.key
	index.Options = indexOptions.SetName(parsedKeys.name)

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

func (mdb *Mdb) NextSerialNumber(colName, key string) (int, error) {
	var no int
	ctx := context.Background()
	err := mdb.Use(colName, "next_number", func(c *mongo.Collection) error {
	again:
		sn := &struct {
			Key string `bson:"_id"`
			No  int    `bson:"no"`
		}{Key: key, No: 1}

		err := c.FindOneAndUpdate(ctx,
			bson.D{{"_id", sn.Key}},
			bson.D{{"$inc", bson.D{{"no", 1}}}},
			options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)).Decode(&sn)
		if err != nil {
			if IsDup(err) {
				goto again
			} else {
				return err
			}
		}
		no = sn.No
		return nil
	})
	return no, err
}

func getMetricKey(defaultMetricKey string, metrics ...string) string {
	if len(metrics) > 0 {
		return strings.Join(metrics, ".")
	}
	return defaultMetricKey
}

type indexKeyInfo struct {
	name string
	key  bson.D
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
				keyInfo.key = append(keyInfo.key, bson.E{Key: "_fts", Value: "text"})
				keyInfo.key = append(keyInfo.key, bson.E{Key: "_ftsx", Value: 1})
				isText = true
			}
		} else if kind != "" {
			keyInfo.key = append(keyInfo.key, bson.E{Key: field, Value: kind})
		} else {
			keyInfo.key = append(keyInfo.key, bson.E{Key: field, Value: order})
		}
	}
	if keyInfo.name == "" {
		return nil, errors.New("invalid index key: no fields provided")
	}
	return &keyInfo, nil
}

// IsDup checks if error is duplicate key error
func IsDup(err error) bool {
	if err == nil {
		return false
	}
	contains := func(s []int, e int) bool {
		for _, a := range s {
			if a == e {
				return true
			}
		}
		return false
	}
	duplicateKeyErrorCodes := []int{11000, 11001, 12582}
	switch e := err.(type) {
	case mongo.WriteException:
		return contains(duplicateKeyErrorCodes, e.WriteErrors[0].Code)
	case mongo.BulkWriteException:
		for _, we := range e.WriteErrors {
			if contains(duplicateKeyErrorCodes, we.Code) {
				return true
			}
		}
	}
	return false
}
