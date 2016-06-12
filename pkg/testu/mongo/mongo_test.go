package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestMongo(t *testing.T) {
	m := New()
	t.Logf("mongo host: %s", m.Host)

	s, err := mgo.Dial(m.Host + "/pero")
	assert.Nil(t, err)
	err = s.Ping()
	assert.Nil(t, err)
	c := s.DB("pero").C("zero")

	err = c.Insert(bson.M{"_id": "0"})
	assert.Nil(t, err)

	m.Stop()
}
