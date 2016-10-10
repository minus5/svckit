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

var db *mdb.Mdb

func main() {
	db = mdb.MustNew("localhost:27017",
		mdb.Name("cacheTest"),
		mdb.CacheCheckpoint(10*time.Second),
		mdb.CacheRoot("./tmp/disk_cache"),
	)
	loop()
	db.Close()
}

func loop() {
	col := "dummys"
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	n := 10
	ds := make([]*Dummy, n)
	for i := 0; i < n; i++ {
		d := &Dummy{}
		if err := db.ReadId(col, i, d); err != nil {
			if err != mdb.ErrNotFound {
				log.Fatal(err)
			}
			d = &Dummy{Id: i}
		}
		ds[i] = d
	}

	for {
		select {
		case <-time.Tick(100 * time.Millisecond):
			i := rand.Intn(n)
			d := ds[i]
			d.Value++
			d.UpdatedAt = time.Now()
			if err := db.SaveId(col, i, d); err != nil {
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
