// +build logs_example

package main

import (
	"bytes"
	"fmt"
	"io"
	"pkg/mdb"
	"time"

	"github.com/minus5/svckit/log"
)

func main() {
	db, err := newDb("localhost:27017", "")
	if err != nil {
		log.Fatal(err)
	}
	fs := db.NewFs("fs")
	//db.createIndexes()

	for i := 0; i < 10; i++ {
		ts := time.Now().Add(time.Duration(-i) * time.Second)
		buf := bytes.NewBufferString(fmt.Sprintf("iso medo u ducan %d %v", i, ts))
		if err := fs.Insert("file", i, ts, buf); err != nil {
			if err == mdb.ErrDuplicate {
				fmt.Printf("duplicate %d\n", i)
				continue
			}
			log.Fatal(err)
		}
	}

	show := func(f io.ReadCloser) error {
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, f)
		if err := f.Close(); err != nil {
			return err
		}
		fmt.Printf("%s\n", buf)
		return nil
	}

	fmt.Println("seek")
	ts := time.Now().Add(-5 * time.Second)
	if err := fs.Seek("file", ts, show); err != nil {
		log.Fatal(err)
	}

	fmt.Println("find")
	if err := fs.Find("file", show); err != nil {
		log.Fatal(err)
	}

	fmt.Println("findId", 7)
	if err := fs.FindId(7, show); err != nil {
		log.Fatal(err)
	}

	fmt.Println("findId", 123)
	if err := fs.FindId(123, show); err != nil {
		if err == mdb.ErrNotFound {
			fmt.Printf("not found\n")
		} else {
			log.Fatal(err)
		}
	}

	if err := fs.Compact("file"); err != nil {
		log.Fatal(err)
	}
}

const dbName = "logs"

type Db struct {
	mdb.Mdb
}

//NewDb - kreira novu konekciju na mongo
func newDb(connStr string, cacheRoot string) (*Db, error) {
	db := &Db{}
	if err := db.Init(connStr,
		mdb.CacheRoot(cacheRoot),
		mdb.EnsureSafe(),
		mdb.Name(dbName)); err != nil {
		return nil, err
	}
	db.Checkpoint()
	return db, nil
}
