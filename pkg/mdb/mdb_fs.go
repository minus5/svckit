package mdb

import (
	"io"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*Fs defines operations for working with append only compactable file system.
Files are gruped by type.
When seeking all files of a type are returned.
Id could be used if it is needed to get a specific file.
*/
type Fs struct {
	name string
	db   *Mdb
}

// Insert file
// typ - type of the file
// id  - colud be omitted if it not required do get by id later
// ts  - timestamp, seek will sort by timestamp
// rdr - content
func (fs *Fs) Insert(typ string, id interface{}, ts time.Time, rdr io.Reader) error {
	return fs.db.UseFs(fs.name, fs.name+"_insert", func(g *mgo.GridFS) error {
		if id != nil {
			_, err := g.OpenId(id)
			if err == nil {
				return ErrDuplicate
			}
		}

		f, err := g.Create(typ)
		if err != nil {
			return translateError(err)
		}
		if id != nil {
			f.SetId(id)
		}
		f.SetUploadDate(ts)
		if _, err := io.Copy(f, rdr); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return translateError(err)
		}
		return nil
	})
}

type seekResult struct {
	Id interface{} `bson:"_id"`
}

// Seek returns all files of a type older then fromTs
func (fs *Fs) Seek(typ string, fromTs time.Time, h func(io.ReadCloser) error) error {
	return fs.db.UseFs(fs.name, fs.name+"_seek", func(g *mgo.GridFS) error {
		i := g.Find(bson.M{"filename": typ, "uploadDate": bson.M{"$gt": fromTs}}).
			Sort("uploadDate").Iter()
		r := seekResult{}
		defer i.Close()
		for i.Next(&r) {
			f, err := g.OpenId(r.Id)
			if err != nil {
				return err
			}
			if err := h(f); err != nil {
				return err
			}
		}
		return nil
	})
}

// FindId returns one file by id
func (fs *Fs) FindId(id interface{}, h func(io.ReadCloser) error) error {
	return fs.db.UseFs(fs.name, fs.name+"_find_id", func(g *mgo.GridFS) error {
		f, err := g.OpenId(id)
		if err != nil {
			return translateError(err)
		}
		if err := h(f); err != nil {
			return translateError(err)
		}
		return nil
	})
}

func translateError(err error) error {
	if mgo.IsDup(err) {
		return ErrDuplicate
	}
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}
	return err
}

// Find retuns last file of a type
func (fs *Fs) Find(typ string, h func(io.ReadCloser) error) error {
	return fs.db.UseFs(fs.name, fs.name+"_find", func(g *mgo.GridFS) error {
		r := seekResult{}
		if err := g.Find(bson.M{"filename": typ}).Sort("-uploadDate").One(&r); err != nil {
			return translateError(err)
		}
		f, err := g.OpenId(r.Id)
		if err != nil {
			return translateError(err)
		}
		if err := h(f); err != nil {
			return translateError(err)
		}
		return nil
	})
}

// Compact deletes all but a last files of a type
func (fs *Fs) Compact(typ string) error {
	return fs.db.UseFs(fs.name, fs.name+"_compact", func(g *mgo.GridFS) error {
		q := g.Find(bson.M{"filename": typ}).Sort("uploadDate")
		r := seekResult{}
		cnt, err := q.Count()
		if err != nil {
			return err
		}
		idx := 0
		i := q.Iter()
		defer i.Close()
		for i.Next(&r) {
			idx++
			if idx >= cnt-1 {
				break
			}
			if err := g.RemoveId(r.Id); err != nil {
				return err
			}
		}
		return nil
	})
}

// Remove deletes all files of a type
func (fs *Fs) Remove(typ string) error {
	return fs.db.UseFs(fs.name, fs.name+"_compact", func(g *mgo.GridFS) error {
		return g.Remove(typ)
	})
}

func (fs *Fs) createIndexes() error {
	return fs.db.Use(fs.name+".files", fs.name+"_indexes", func(c *mgo.Collection) error {
		if err := c.EnsureIndex(mgo.Index{
			Key: []string{"filename", "uploadDate"},
		}); err != nil {
			return err
		}
		return nil
	})
}
