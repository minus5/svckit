package util

import (
	"compress/gzip"
	"io/ioutil"
	"bytes"
)

//Gzip - compess input
func Gzip(data []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

//GzipStr - cast for me
func GzipStr(data string) string {
	return string(Gzip([]byte(data)))
}

//Gunzip - decompress data
func Gunzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(data)
	r, err := gzip.NewReader(&buf)
	if err != nil {
		return nil, err
	}
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return out, nil
}

//GunzipStr - cast for me
func GunzipStr(data string) (string, error) {
	ret, err := Gunzip([]byte(data))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func IsGziped(buf []byte) bool {
	if len(buf) > 2 {
		return buf[0] == 0x1f && buf[1] == 0x8b
	}
	return false
}
