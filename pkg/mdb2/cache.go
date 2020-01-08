package mdb2

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/minus5/svckit/log"

	"go.mongodb.org/mongo-driver/bson"
)

type cache struct {
	mdb *Mdb
	m   map[string]*cacheItem
	sync.Mutex
}

func newCache(mdb *Mdb) (*cache, error) {
	if err := os.MkdirAll(mdb.cacheDir, os.ModePerm); err != nil {
		return nil, err
	}
	c := &cache{
		m:   make(map[string]*cacheItem),
		mdb: mdb,
	}
	c.init()
	return c, nil
}

func (c *cache) init() {
	type obj struct {
		Id interface{} `bson:"_id"`
	}

	// check if anything is left in disk cache
	files, _ := ioutil.ReadDir(c.mdb.cacheDir)
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

		raw, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", c.mdb.cacheDir, f.Name()))
		if err != nil {
			log.Error(err)
		}
		// deserialize to get Id in appropriate type
		o := &obj{}
		if err := bson.Unmarshal(raw, o); err == nil {
			id = o.Id
		}

		i := newCacheItem(col, id, raw, c.mdb.cacheDir)
		c.m[i.key] = i
	}
}

func (c *cache) add(col string, id interface{}, o interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := make([]byte, 20480)
			stackSize := runtime.Stack(stackTrace, true)
			log.S("id", fmt.Sprintf("%v", id)).
				S("col", col).
				S("panic", fmt.Sprintf("%v", r)).
				I("stack_size", stackSize).
				S("stack_trace", string(stackTrace)).
				ErrorS("Recover from panic")
		}
	}()
	raw, err := bson.Marshal(o)
	if err != nil {
		return err
	}
	i := newCacheItem(col, id, raw, c.mdb.cacheDir)
	c.Lock()
	c.m[i.key] = i
	c.Unlock()
	return c.save(i)
}

func (c *cache) find(col string, id interface{}) (*cacheItem, bool) {
	key := fmt.Sprintf("%s.%v", col, id)
	c.Lock()
	defer c.Unlock()
	i, ok := c.m[key]
	return i, ok
}

func (c *cache) remove(col string, id interface{}) {
	key := fmt.Sprintf("%s.%v", col, id)
	c.Lock()
	defer c.Unlock()
	if i, ok := c.m[key]; ok {
		os.Remove(i.fn)
	}

}

func (c *cache) save(i *cacheItem) error {
	return ioutil.WriteFile(i.fn, i.raw, os.ModePerm)
}

// purge saves all cached documents into database
// and removes them from cache
func (c *cache) purge() {
	m := make(map[string]*cacheItem)
	c.Lock()
	for k, i := range c.m {
		m[k] = i
	}
	c.Unlock()

	for k, i := range m {
		c.Lock()
		delete(c.m, k)
		c.Unlock()
		err := c.mdb.saveId(i.col, "saveId", i.id, i.o())
		if err != nil {
			log.S("col", i.col).S("id", fmt.Sprintf("%v", i.id)).Error(err)
		}
		c.Lock()
		if _, found := c.m[k]; found {
			c.Unlock()
			continue
		}
		if err == nil {
			// remove from disk
			err2 := os.Remove(i.fn)
			if err2 != nil {
				log.Error(err2)
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

// object to save into mongo
func (i *cacheItem) o() *bson.Raw {
	r := bson.Raw(i.raw)
	return &r
}

func (i *cacheItem) unmarshal(o interface{}) error {
	return bson.Unmarshal(i.raw, o)
}
