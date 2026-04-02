package compress

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	// Kreiraj test gzip data od 3MB za BenchmarkGzip
	createTestData3MB()
	// Kreiraj tesni cache za BenchmarkGzip, cache ce imati 5 entry-a
	// Ideja je da probam napraviti profile potrosnje i oslobadjanja memorije
	createTestCache(5)

	// Run tests
	os.Exit(m.Run())
}

func TestGunzip(t *testing.T) {
	uncompressedStr := "iso medo u ducan"
	compressed := []byte{31, 139, 8, 0, 18, 187, 71, 83, 0, 3, 203, 44, 206, 87, 200, 77, 77, 201, 87, 40, 85, 72, 41, 77, 78, 204, 3, 0, 31, 207, 155, 86, 16, 0, 0, 0}
	u, err := Gunzip(compressed)
	assert.Nil(t, err)
	assert.Equal(t, uncompressedStr, string(u))

	c := Gzip([]byte(uncompressedStr))
	u, err = Gunzip(c)
	assert.Nil(t, err)
	assert.Equal(t, uncompressedStr, string(u))

	assert.True(t, IsGziped(compressed))
	assert.False(t, IsGziped([]byte(uncompressedStr)))

	c2 := GzipStr(uncompressedStr)
	s2, err := GunzipStr(string(c2))
	assert.Nil(t, err)
	assert.Equal(t, uncompressedStr, s2)

	u2, err := GunzipIf(c)
	assert.Nil(t, err)
	assert.Equal(t, uncompressedStr, string(u2))

	u3, err := GunzipIf([]byte(uncompressedStr))
	assert.Nil(t, err)
	assert.Equal(t, uncompressedStr, string(u3))
}

func TestGzipper(t *testing.T) {
	gz := NewGzipper()
	ts := "iso mendo u ducan nije rekao dobar dan "
	gzd, err := gz.Gzip([]byte(ts))
	assert.NoError(t, err)
	assert.NotNil(t, gzd)
	assert.NotEmpty(t, gzd)

	// GUnzip za provjeru
	ungzd, err := Gunzip(gzd)
	assert.NoError(t, err)
	assert.NotNil(t, ungzd)
	assert.NotEmpty(t, ungzd)

	// Da li je isto sto smo gzipali
	assert.Equal(t, ts, string(ungzd))
}

func TestGzipperMulti(t *testing.T) {
	var wg sync.WaitGroup
	gz := NewGzipper()

	fn := func(ts string) {
		gzd, err := gz.Gzip([]byte(ts))
		assert.NoError(t, err)
		assert.NotNil(t, gzd)
		assert.NotEmpty(t, gzd)

		// GUnzip za provjeru
		ungzd, err := Gunzip(gzd)
		assert.NoError(t, err)
		assert.NotNil(t, ungzd)
		assert.NotEmpty(t, ungzd)

		// Da li je isto sto smo gzipali
		assert.Equal(t, ts, string(ungzd))
		t.Log(string(ungzd))
		wg.Done()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		ts := fmt.Sprintf("iso medo u ducan nije rekao dobar dan %d", i)
		go fn(ts)
	}
	wg.Wait()
}

// testni podaci za gzipanje 3mb+
var data3MB []byte

// createTestData3MB kreira testne podatke za gzipanje
func createTestData3MB() {
	var bb bytes.Buffer
	i := 0
	ts := "iso medo u ducan nije rekao dobar dan "
	for bb.Len() < 3*1024*1024 {
		bb.WriteString(fmt.Sprintf(`{no:%d, msg:"%s"}`, i, ts))
		ts = ts[1:] + ts[0:1]
		i++
	}
	data3MB = bb.Bytes()
}

// CompressedHolder Struktura koja sadrzi kompresirane podatke
type CompressedHolder struct {
	CData []byte
}

// cachePos trenutna pozicija u cache-u gdje cemo za test dodavati kompresirane podatke u memoriju
var cachePos int

// cache mapa koja cuva pointere na strukture kompresiranih podataka
var cache map[int]*CompressedHolder

// createTestCache kreira mapu za cache neke velicine da pisemo po njoj u krug
func createTestCache(maxCacheSize int) {
	cachePos = 0
	cache = make(map[int]*CompressedHolder, maxCacheSize)
}

// cacheAdd dodaje u cache za test pa da probamo pratiti alokaciju i oslobadjanje memorije
func cacheAdd(b []byte) {
	cache[cachePos] = &CompressedHolder{
		CData: b,
	}
	cachePos++
	// Resetiraj cache poziciju da pisemo u krug oviso o velicini cache-a
	if cachePos >= len(cache) {
		cachePos = 0
	}
}

func TestZstd(t *testing.T) {
	input := []byte("iso medo u ducan nije rekao dobar dan")

	compressed := Zstd(input)
	assert.True(t, IsZstd(compressed))
	assert.False(t, IsZstd(input))

	out, err := Unzstd(compressed)
	assert.Nil(t, err)
	assert.Equal(t, input, out)
}

func TestSnappy(t *testing.T) {
	input := []byte("iso medo u ducan nije rekao dobar dan")

	compressed := Snappy(input)
	assert.True(t, IsSnappy(compressed))
	assert.False(t, IsSnappy(input))

	out, err := Unsnappy(compressed)
	assert.Nil(t, err)
	assert.Equal(t, input, out)
}

func TestLz4(t *testing.T) {
	input := []byte("iso medo u ducan nije rekao dobar dan")

	compressed := Lz4(input)
	assert.True(t, IsLz4(compressed))
	assert.False(t, IsLz4(input))

	out, err := Unlz4(compressed)
	assert.Nil(t, err)
	assert.Equal(t, input, out)
}

func TestDetect(t *testing.T) {
	input := []byte("iso medo u ducan nije rekao dobar dan")

	assert.Equal(t, EncodingGzip, Detect(Gzip(input)))
	assert.Equal(t, EncodingZstd, Detect(Zstd(input)))
	assert.Equal(t, EncodingSnappy, Detect(Snappy(input)))
	assert.Equal(t, EncodingLz4, Detect(Lz4(input)))
	assert.Equal(t, EncodingNone, Detect(input))
}

func TestDecompressIf(t *testing.T) {
	input := []byte("iso medo u ducan nije rekao dobar dan")

	cases := []struct {
		name     string
		compress func([]byte) []byte
		encoding Encoding
	}{
		{"gzip", Gzip, EncodingGzip},
		{"zstd", Zstd, EncodingZstd},
		{"snappy", Snappy, EncodingSnappy},
		{"lz4", Lz4, EncodingLz4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, enc, err := DecompressIf(tc.compress(input))
			assert.Nil(t, err)
			assert.Equal(t, input, out)
			assert.Equal(t, tc.encoding, enc)
		})
	}

	// uncompressed passthrough
	out, enc, err := DecompressIf(input)
	assert.Nil(t, err)
	assert.Equal(t, input, out)
	assert.Equal(t, EncodingNone, enc)
}

// BenchmarkGzip benchmark za gzip da vidim potrosnju memorije
// Pokrecem da radi samo BenchmarkGzip
// go test -v -run NOTest -benchmem -benchtime 10s -bench BenchmarkGzip -memprofile=mem0.out
// Memory profile radim kasnije sa:
// go tool pprof --alloc_space util.test mem0.out
func BenchmarkGzip(b *testing.B) {
	// Zelim alokacije memorije
	b.ReportAllocs()
	// Resetiraj timer mjerimo vrijeme
	b.ResetTimer()
	// Radi bench
	for i := 0; i < b.N; i++ {
		cacheAdd(Gzip(data3MB))
	}
}

// BenchmarkNewGzipper benchmark za gzip da vidim potrosnju memorije
// Pokrecem da radi samo BenchmarkGzip
// go test -v -run NOTest -benchmem -benchtime 10s -bench BenchmarkNewGzipper -memprofile=mem1.out
// Memory profile radim kasnije sa:
// go tool pprof --alloc_space util.test mem1.out
func BenchmarkNewGzipper(b *testing.B) {
	gz := NewGzipper()
	// Zelim alokacije memorije
	b.ReportAllocs()
	// Resetiraj timer mjerimo vrijeme
	b.ResetTimer()
	// Radi bench
	for i := 0; i < b.N; i++ {
		gzd, _ := gz.Gzip(data3MB)
		cacheAdd(gzd)
	}
}
