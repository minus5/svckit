package util

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"sync"
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
	defer r.Close()
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GunzipIf gunzipa ako je data gzipan
func GunzipIf(data []byte) ([]byte, error) {
	if IsGziped(data) {
		return Gunzip(data)
	}
	return data, nil
}

//GunzipStr - cast for me
func GunzipStr(data string) (string, error) {
	ret, err := Gunzip([]byte(data))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// IsGziped provjerava da li je buffer gzipan
func IsGziped(buf []byte) bool {
	if len(buf) > 2 {
		return buf[0] == 0x1f && buf[1] == 0x8b
	}
	return false
}

// Gzipper koristi jedan gzip writer, namijenjen je da se koristi single threaded
// jer uvijek koristi isti buffer za pisanje kompresiranih podataka
type Gzipper struct {
	b bytes.Buffer // buffer za kompresirane podatke
	w *gzip.Writer // writer koji obavlja gzipanje
	sync.Mutex
}

// NewGzipper kreira novi Gzipper
func NewGzipper() *Gzipper {
	g := &Gzipper{}
	g.w = gzip.NewWriter(&g.b)
	return g
}

// Gzip kompresira podatke i vraca kompresiranu kopiju kako bi se buffer
// za kompresiranje mogao ponovo korisiti
func (g *Gzipper) Gzip(data []byte) ([]byte, error) {
	g.Lock()
	defer g.Unlock()
	g.w.Reset(&g.b)
	g.b.Reset()
	if _, err := g.w.Write(data); nil != err {
		return nil, err
	}
	if err := g.w.Close(); nil != err {
		return nil, err
	}
	c := make([]byte, g.b.Len())
	copy(c, g.b.Bytes())
	return c, nil
}
