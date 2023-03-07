package mdb2

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/pkg/testu2/mongo"
	"github.com/stretchr/testify/assert"
)

type obj struct {
	Id int `bson:"_id"`
}

type obj2 struct {
	Id interface{} `bson:"_id"`
}

type db2 struct {
	Mdb
}

const (
	testDbName         = "test_mdb"
	testCollectionName = "col"
	testTopic          = "test.topic"
)

var mongoSrv *mongo.Mongo
var db *db2

func TestMain(m *testing.M) {
	// starts test mongo server
	mongoSrv = mongo.New()
	log.SetOutput(ioutil.Discard)

	os.Exit(m.Run())
}

func setup(t *testing.T) {
	db = &db2{}
	if err := db.Init(DefaultConnStr(), MajoritySafe(), Name(testDbName)); err != nil {
		log.Fatalf("failed to open connection: %s", err)
		return
	}
}

var testCacheDir = "./tmp/cacheDir"

func TestCacheAdd(t *testing.T) {
	db := &Mdb{name: "dbName", cacheDir: testCacheDir}
	c, err := newCache(db)
	assert.Nil(t, err)

	err = c.add("obj", 1, &obj{Id: 1})
	assert.Nil(t, err)
	err = c.add("obj", 2, &obj{Id: 2})
	assert.Nil(t, err)
	err = c.add("obj", 3, &obj{Id: 3})
	assert.Nil(t, err)

	ls := ls(testCacheDir)
	assert.Equal(t, []string{"obj.1", "obj.2", "obj.3"}, ls)
	t.Logf("%v", ls)

	c2, err := newCache(db)
	assert.Nil(t, err)
	assert.Len(t, c2.m, 3)

	i := c2.m["obj.3"]
	assert.Equal(t, int32(3), i.id)
}

func ls(dir string) []string {
	var ls []string
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ls = append(ls, f.Name())
	}
	return ls
}

type ts struct {
	Id    interface{} `bson:"_id,omitemtpy"`
	Value int         `bson:"value"`
}

func TestSerializeJson(t *testing.T) {
	t1 := &ts{Id: 1}
	b, err := json.Marshal(t1)
	assert.Nil(t, err)
	fmt.Printf("%s\n", b)

	t2 := &ts{}
	err = json.Unmarshal(b, t2)
	assert.Nil(t, err)
	fmt.Printf("%#v %s\n", t2, reflect.TypeOf(t2.Id))
}

func TestSerializeGob(t *testing.T) {
	t1 := &ts{Id: 1}

	b := new(bytes.Buffer)
	ge := gob.NewEncoder(b)
	err := ge.Encode(t1)
	assert.Nil(t, err)

	t2 := &ts{}
	gd := gob.NewDecoder(bytes.NewBuffer(b.Bytes()))
	err = gd.Decode(t2)
	assert.Nil(t, err)

	fmt.Printf("%#v %s\n", t2, reflect.TypeOf(t2.Id))
}

func TestReadId(t *testing.T) {
	t1 := &ts{Id: 1}
	assert.Equal(t, 1, _id(t1))
}

func _id(o interface{}) interface{} {
	s := reflect.ValueOf(o).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		t := typeOfT.Field(i).Tag
		f := s.Field(i)
		if strings.HasPrefix(t.Get("bson"), "_id") {
			return f.Interface()
		}
		fmt.Printf("%d: %s %s = %v tag: %s %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface(), t, t.Get("bson"))
	}
	return nil
}

func TestMongoSerde(t *testing.T) {
	setup(t)
	m := map[string]interface{}{
		"a": 12345,
		"b": map[string]interface{}{
			"i":   12345678901235,
			"i64": int64(12345678901235),
			"i32": int32(1234567),
		},
		"c": 12345678901234,
	}
	db.SaveId(testCollectionName, 2513, m)

	var res map[string]interface{}
	db.ReadId(testCollectionName, 2513, &res)

	assert.Equal(t, m["a"], res["a"])
	assert.Equal(t, int64(m["b"].(map[string]interface{})["i"].(int)), res["b"].(map[string]interface{})["i"])
	assert.Equal(t, m["b"].(map[string]interface{})["i64"].(int64), res["b"].(map[string]interface{})["i64"])
	assert.Equal(t, int(m["b"].(map[string]interface{})["i32"].(int32)), res["b"].(map[string]interface{})["i32"])
	assert.Equal(t, int64(m["c"].(int)), res["c"])
}
