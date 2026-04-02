package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"

	"github.com/klauspost/compress/snappy"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
)

// Encoding is the name of a compression algorithm as used in message headers.
type Encoding string

const (
	EncodingNone   Encoding = ""
	EncodingGzip   Encoding = "gzip"
	EncodingZstd   Encoding = "zstd"
	EncodingSnappy Encoding = "snappy"
	EncodingLz4    Encoding = "lz4"
)

var (
	zstdEncoder *zstd.Encoder
	zstdDecoder *zstd.Decoder
)

func init() {
	var err error
	zstdEncoder, err = zstd.NewWriter(nil)
	if err != nil {
		panic(err)
	}
	zstdDecoder, err = zstd.NewReader(nil)
	if err != nil {
		panic(err)
	}
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
	out, err := io.ReadAll(r)
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

// Zstd compresses data with zstandard.
func Zstd(data []byte) []byte {
	return zstdEncoder.EncodeAll(data, nil)
}

// Unzstd decompresses zstandard data.
func Unzstd(data []byte) ([]byte, error) {
	return zstdDecoder.DecodeAll(data, nil)
}

// IsZstd reports whether buf starts with the zstandard frame magic.
func IsZstd(buf []byte) bool {
	return len(buf) >= 4 &&
		buf[0] == 0x28 && buf[1] == 0xB5 && buf[2] == 0x2F && buf[3] == 0xFD
}

// Snappy compresses data using the snappy streaming (framing) format.
func Snappy(data []byte) []byte {
	var b bytes.Buffer
	w := snappy.NewBufferedWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

// Unsnappy decompresses snappy streaming-format data.
func Unsnappy(data []byte) ([]byte, error) {
	r := snappy.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}

// IsSnappy reports whether buf starts with the snappy framing-format stream identifier.
func IsSnappy(buf []byte) bool {
	// Stream identifier chunk: 0xff followed by chunk length 0x060000 and "sNaPpY"
	return len(buf) >= 10 &&
		buf[0] == 0xff &&
		buf[1] == 0x06 && buf[2] == 0x00 && buf[3] == 0x00 &&
		buf[4] == 0x73 && buf[5] == 0x4e && buf[6] == 0x61 &&
		buf[7] == 0x50 && buf[8] == 0x70 && buf[9] == 0x59
}

// Lz4 compresses data using the lz4 frame format.
func Lz4(data []byte) []byte {
	var b bytes.Buffer
	w := lz4.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

// Unlz4 decompresses lz4 frame-format data.
func Unlz4(data []byte) ([]byte, error) {
	r := lz4.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}

// IsLz4 reports whether buf starts with the lz4 frame magic number.
func IsLz4(buf []byte) bool {
	return len(buf) >= 4 &&
		buf[0] == 0x04 && buf[1] == 0x22 && buf[2] == 0x4d && buf[3] == 0x18
}

// Detect returns the encoding of buf or EncodingNone if unrecognized.
func Detect(buf []byte) Encoding {
	switch {
	case IsGziped(buf):
		return EncodingGzip
	case IsZstd(buf):
		return EncodingZstd
	case IsSnappy(buf):
		return EncodingSnappy
	case IsLz4(buf):
		return EncodingLz4
	}
	return EncodingNone
}

// DecompressIf decompresses buf if a known compression format is detected.
// Returns the (possibly decompressed) data, the detected encoding, and any error.
func DecompressIf(data []byte) ([]byte, Encoding, error) {
	switch {
	case IsGziped(data):
		out, err := Gunzip(data)
		return out, EncodingGzip, err
	case IsZstd(data):
		out, err := Unzstd(data)
		return out, EncodingZstd, err
	case IsSnappy(data):
		out, err := Unsnappy(data)
		return out, EncodingSnappy, err
	case IsLz4(data):
		out, err := Unlz4(data)
		return out, EncodingLz4, err
	}
	return data, EncodingNone, nil
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
