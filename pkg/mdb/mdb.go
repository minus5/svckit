package mdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ErrNotFound raised when record is not found in db
var ErrNotFound = errors.New("not found")

type cache struct {
	db *Mdb
	m  map[string]*cacheItem
	sync.Mutex
}

func newCache(db *Mdb) (*cache, error) {
	if err := os.MkdirAll(db.cacheDir, os.ModePerm); err != nil {
		return nil, err
	}
	c := &cache{
		m:  make(map[string]*cacheItem),
		db: db,
	}
	c.init()
	return c, nil
}

func (c *cache) init() {
	type obj struct {
		Id interface{} `bson:"_id"`
	}

	// check if anything is left into disk cache
	files, _ := ioutil.ReadDir(c.db.cacheDir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		p := strings.Split(f.Name(), ".")
		if len(p) != 2 {
			continue
		}
		var id interface{}
		col := p[0]
		id = p[1]

		raw, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", c.db.cacheDir, f.Name()))
		if err != nil {
			log.Error(err)
		}
		// deserialize to get Id in appropirate type
		o := &obj{}
		if err := bson.Unmarshal(raw, o); err == nil {
			id = o.Id
		}

		i := newCacheItem(col, id, raw, c.db.cacheDir)
		c.m[i.key] = i
	}
}

// add item to cache
func (c *cache) add(col string, id interface{}, o interface{}) error {
	raw, err := bson.Marshal(o)
	if err != nil {
		return err
	}
	i := newCacheItem(col, id, raw, c.db.cacheDir)
	c.Lock()
	c.m[i.key] = i
	c.Unlock()
	return c.save(i)
}

// find into cache
func (c *cache) find(col string, id interface{}) (*cacheItem, bool) {
	key := fmt.Sprintf("%s.%v", col, id)
	c.Lock()
	defer c.Unlock()
	i, ok := c.m[key]
	return i, ok
}

func newCacheItem(col string, id interface{}, raw []byte, cacheDir string) *cacheItem {
	key := fmt.Sprintf("%s.%v", col, id)
	i := &cacheItem{
		col: col,
		id:  id,
		raw: raw,
		key: key,
		fn:  fmt.Sprintf("%s/%s", cacheDir, key),
	}
	return i
}

// save cacheItem to disk
func (c *cache) save(i *cacheItem) error {
	return ioutil.WriteFile(i.fn, i.raw, os.ModePerm)
}

// purge saves all cached item into database
// and removes them from cache
func (c *cache) purge() {
	for k, i := range c.m {
		c.Lock()
		// delete from cache
		delete(c.m, k)
		c.Unlock()
		// save to db
		err := c.db.saveId(i.col, i.id, i.o())
		if err != nil {
			log.Error(err)
		}
		c.Lock()
		// if new exists into cache do nothing
		if _, found := c.m[k]; found {
			c.Unlock()
			continue
		}
		if err == nil {
			// remove from disk
			err2 := os.Remove(i.fn)
			if err2 != nil {
				log.Error(err)
			}
		} else {
			// in case of error return to cache
			c.m[k] = i
		}
		c.Unlock()
	}
}

type cacheItem struct {
	col string
	id  interface{}
	raw []byte
	key string
	fn  string
}

// o object to save into mongo
// Mongo understands bson.Raw type.
func (i *cacheItem) o() *bson.Raw {
	return &bson.Raw{Data: i.raw}
}

// unmarshal cacheItem to type
func (i *cacheItem) unmarshal(o interface{}) error {
	return bson.Unmarshal(i.raw, o)
}

// Mdb konekacija i operacije s bazom
type Mdb struct {
	name         string
	session      *mgo.Session
	cacheDir     string
	checkPointIn time.Duration
	cache        *cache
}

// MustNew raises fatal is unable to connect to mongo
func MustNew(connStr string, opts ...func(db *Mdb)) *Mdb {
	db, err := NewDb(connStr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Name sets mongo database name, default is application name
func Name(n string) func(db *Mdb) {
	return func(db *Mdb) {
		db.name = n
	}
}

// CacheRoot sets disk cache root directory
func CacheRoot(d string) func(db *Mdb) {
	return func(db *Mdb) {
		if d != "" {
			db.cacheDir = fmt.Sprintf("%s/%s", d, db.name)
		}
	}
}

// CacheCheckpoint sets checkpoint interval.
// When all cached items are flushed into mongo.
func CacheCheckpoint(d time.Duration) func(db *Mdb) {
	return func(db *Mdb) {
		db.checkPointIn = d
	}
}

// NewDb creates new Db
// Connects to mongo, initializes cache, starts checkpoint loop.
func NewDb(connStr string, opts ...func(db *Mdb)) (*Mdb, error) {
	db := &Mdb{}
	if err := db.Init(connStr, opts...); err != nil {
		return nil, err
	}
	return db, nil
}

// Init initializes new Mdb
// Connects to mongo, initializes cache, starts checkpoint loop.
func (db *Mdb) Init(connStr string, opts ...func(db *Mdb)) error {
	db.checkpoint()
	s, err := mgo.Dial(connStr)
	if err != nil {
		return err
	}
	s.SetMode(mgo.Eventual, false)
	s.SetSafe(nil)
	db.session = s
	// defaults
	db.name = env.AppName()
	db.checkPointIn = 30 * time.Second
	// apply options
	for _, opt := range opts {
		opt(db)
	}
	if db.cacheDir != "" {
		db.cache, err = newCache(db)
		if err != nil {
			return err
		}
		go db.loop()
	}
	return nil
}

// Ping returns true if mongo is available
func (db *Mdb) Ping() bool {
	s := db.session.Copy()
	defer s.Close()
	return s.Ping() == nil
}

func (db *Mdb) loop() {
	for {
		select {
		case <-time.Tick(db.checkPointIn):
			db.checkpoint()
		}
	}
}

func (db *Mdb) checkpoint() {
	if db.cache != nil {
		db.cache.purge()
	}
}

// Close starts clean exit
func (db *Mdb) Close() {
	db.checkpoint()
}

// Checkpoint flush caches
func (db *Mdb) Checkpoint() {
	db.checkpoint()
}

func (db *Mdb) Use(col string, metricKey string, handler func(*mgo.Collection) error) error {
	s := db.session.Copy()
	defer s.Close()
	c := s.DB(db.name).C(col)
	var err error
	metric.Timing("db."+metricKey, func() {
		err = handler(c)
	})
	return err
}

// SaveId stores document to cache
// or directly to mongo if cache is not enabled
func (db *Mdb) SaveId(col string, id interface{}, o interface{}) error {
	if db.cache != nil {
		return db.cache.add(col, id, o)
	}
	return db.saveId(col, id, o)
}

func (db *Mdb) saveId(col string, id interface{}, o interface{}) error {
	return db.Use(col, "saveId", func(c *mgo.Collection) error {
		_, err := c.UpsertId(id, o)
		return err
	})
}

// ReadId retruns document from cache or mongo
func (db *Mdb) ReadId(col string, id interface{}, o interface{}) error {
	if db.cache != nil {
		// try to find in cache
		if i, ok := db.cache.find(col, id); ok {
			log.S("col", col).S("id", fmt.Sprintf("%v", id)).Info("ReadId from cache")
			return i.unmarshal(o)
		}
	}
	// go to mongo
	err := db.Use(col, "readId", func(c *mgo.Collection) error {
		err := c.FindId(id).One(o)
		if err == mgo.ErrNotFound {
			return ErrNotFound
		}
		return err
	})
	return err
}
