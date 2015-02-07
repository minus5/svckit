package util

import (
	"github.com/stretchrcom/testify/assert"
	"testing"
	)

func TestGunzip(t *testing.T) {
	uncompressedStr := "iso medo u ducan"
	compressed := []byte{31,139,8,0,18,187,71,83,0,3,203,44,206,87,200,77,77,201,87,40,85,72,41,77,78,204,3,0,31,207,155,86,16,0,0,0}
	u, err := Gunzip(compressed)
	assert.Nil(t, err)
	assert.Equal(t, uncompressedStr, string(u)) 

	c := Gzip([]byte(uncompressedStr))
	u, err = Gunzip(c)
	assert.Nil(t, err)
	assert.Equal(t, uncompressedStr, string(u)) 
}
