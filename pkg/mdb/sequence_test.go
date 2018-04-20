package mdb

import (
	"pkg/testu/mongo"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequence(t *testing.T) {
	srv := mongo.New()
	db, err := NewDb(srv.Host)
	assert.Nil(t, err)

	s, err := db.GetSequence("seq1", 4)
	assert.Nil(t, err)

	next := func(expected int) {
		i, err := s.Next()
		assert.Nil(t, err)
		assert.Equal(t, uint64(expected), i)
	}

	find := func(expected int) {
		i, err := s.find()
		assert.Nil(t, err)
		assert.Equal(t, uint64(expected), i)
	}

	find(4)
	next(1)
	next(2)
	find(4)
	next(3)
	next(4)
	find(8)
	next(5)
	next(6)
	next(7)
	next(8)
	find(12)

	// dodajem jos jednu na isti key da napravim nered
	s1, err := db.GetSequence("seq1", 2)
	assert.Nil(t, err)
	i, err := s1.Next()
	assert.Nil(t, err)
	assert.Equal(t, uint64(12), i)

	find(14)
	next(9)
	next(10)
	next(11)
	_, err = s.Next()
	assert.NotNil(t, err)

	i, err = s1.Next()
	assert.Nil(t, err)
	assert.Equal(t, uint64(13), i)
	assert.Equal(t, uint64(14), s1.leased)

	i, err = s1.Next()
	assert.Nil(t, err)
	assert.Equal(t, uint64(14), i)
	assert.Equal(t, uint64(16), s1.leased)

	err = s1.Release()
	assert.Nil(t, err)
	find(15)

	db.Close()
	srv.Stop()
}
