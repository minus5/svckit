package amp

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// benchPayload is a realistic amp payload: a large, highly repetitive JSON
// snapshot, the shape that clears compressionLenLimit in production.
func benchPayload(events int) []byte {
	var b strings.Builder
	b.WriteString(`{"events":[`)
	for i := range events {
		fmt.Fprintf(&b, `{"id":%d,"name":"event %d","odds":1.%03d,"type":"m"},`, i, i, i%1000)
	}
	b.WriteString(`{"end":0}],"rid":"_auto_:2p1c8hv3btvif6"}`)
	return []byte(b.String())
}

func BenchmarkDeflate(b *testing.B) {
	for _, events := range []int{100, 2000, 20000} {
		src := benchPayload(events)
		b.Run(fmt.Sprintf("src=%dB", len(src)), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(src)))
			for b.Loop() {
				sink = deflate(src)
			}
		})
	}
}

func BenchmarkUndeflate(b *testing.B) {
	for _, events := range []int{100, 2000, 20000} {
		src := benchPayload(events)
		enc := deflate(src)
		b.Run(fmt.Sprintf("src=%dB", len(src)), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(enc)))
			for b.Loop() {
				sink = Undeflate(enc)
			}
		})
	}
}

// BenchmarkDeflateParallel is the one that matters operationally: every
// websocket writer compresses its own message, so the hot path is concurrent.
func BenchmarkDeflateParallel(b *testing.B) {
	src := benchPayload(2000)
	b.ReportAllocs()
	b.SetBytes(int64(len(src)))
	b.RunParallel(func(pb *testing.PB) {
		// Local, not the package-level sink: concurrent workers writing one
		// shared slice header is a data race of the benchmark's own making, and
		// it contends on a location that has nothing to do with the pooling
		// being measured.
		var local []byte
		for pb.Next() {
			local = deflate(src)
		}
		runtime.KeepAlive(local)
	})
}

var sink []byte

// BenchmarkMarshalDeflate measures the whole production path, cache miss each
// time (a fresh Msg per iteration, as in a real publish).
func BenchmarkMarshalDeflate(b *testing.B) {
	body := benchPayload(2000)
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		m := &Msg{Type: Publish, URI: "sport/events", body: body}
		sink, _ = m.MarshalDeflate()
	}
}
