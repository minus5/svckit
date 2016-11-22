// +build main

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"pkg/mdb"
	"syscall"
	"time"

	"github.com/minus5/svckit/log"
)

// Dummy ...
type Dummy struct {
	Id        int `bson:"_id"`
	Value     int
	UpdatedAt time.Time
}

var db *Db

func main() {
	err := db.Init("localhost:27017",
		mdb.Name("cacheTest"),
		mdb.CacheCheckpoint(10*time.Second),
		mdb.CacheRoot("./tmp/disk_cache"),
	)
	if err != nil {
		log.Fatal(err)
	}
	loop()
	db.Close()
}

func loop() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	n := 10
	ds := make([]*Dummy, n)
	for i := 0; i < n; i++ {
		d := db.ReadDummy(i)
		ds[i] = d
	}

	for {
		select {
		case <-time.Tick(100 * time.Millisecond):
			i := rand.Intn(n)
			d := ds[i]
			d.Value++
			d.UpdatedAt = time.Now()
			if err := db.SaveDummy(d); err != nil {
				log.Fatal(err)
			}
			if i == 0 {
				fmt.Printf("%d  %d  %v\n", i, d.Value, d.UpdatedAt)
			}
		case <-c:
			return
		}
	}
}

var dummysCol = "dummys"

type Db struct {
	mdb.Mdb
}

func (db *Db) SaveDummy(d *Dummy) error {
	return db.SaveId(dummysCol, d.Id, d)
}

func (db *Db) ReadDummy(id int) *Dummy {
	d := &Dummy{}
	if err := db.ReadId(dummysCol, id, d); err != nil {
		if err != mdb.ErrNotFound {
			log.Fatal(err)
		}
		d = &Dummy{Id: id}
	}
	return d
}
