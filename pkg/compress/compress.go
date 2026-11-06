package compress

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// Encoding is the value of the "encoding" header of a backend message.
type Encoding string

const (
	EncodingNone Encoding = ""
	EncodingGzip Encoding = "gzip"
	EncodingZstd Encoding = "zstd"
)

// Shared: EncodeAll and DecodeAll are safe for concurrent use.
var (
	zstdEncoder = mustNewZstdEncoder()
	zstdDecoder = mustNewZstdDecoder()
)

func mustNewZstdEncoder() *zstd.Encoder {
	e, err := zstd.NewWriter(nil)
	if err != nil {
		panic(err)
	}
	return e
}

func mustNewZstdDecoder() *zstd.Decoder {
	d, err := zstd.NewReader(nil)
	if err != nil {
		panic(err)
	}
	return d
}

func Zstd(data []byte) []byte {
	return zstdEncoder.EncodeAll(data, nil)
}

func Unzstd(data []byte) ([]byte, error) {
	return zstdDecoder.DecodeAll(data, nil)
}

func IsZstd(buf []byte) bool {
	return len(buf) >= 4 &&
		buf[0] == 0x28 && buf[1] == 0xB5 && buf[2] == 0x2F && buf[3] == 0xFD
}

func Detect(buf []byte) Encoding {
	switch {
	case IsGziped(buf):
		return EncodingGzip
	case IsZstd(buf):
		return EncodingZstd
	}
	return EncodingNone
}

// DecompressIf is GunzipIf extended to zstd: on a corrupt body it returns the
// error with whatever the decoder produced.
func DecompressIf(data []byte) ([]byte, Encoding, error) {
	switch enc := Detect(data); enc {
	case EncodingGzip:
		out, err := Gunzip(data)
		return out, enc, err
	case EncodingZstd:
		out, err := Unzstd(data)
		return out, enc, err
	}
	return data, EncodingNone, nil
}

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
