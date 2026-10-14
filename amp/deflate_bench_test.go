package amp

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// benchPayload is a large repetitive JSON snapshot.
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

func BenchmarkDeflateParallel(b *testing.B) {
	src := benchPayload(2000)
	b.ReportAllocs()
	b.SetBytes(int64(len(src)))
	b.RunParallel(func(pb *testing.PB) {
		// Local sink: a shared one would race.
		var local []byte
		for pb.Next() {
			local = deflate(src)
		}
		runtime.KeepAlive(local)
	})
}

var sink []byte

// BenchmarkMarshalDeflate uses a fresh Msg per iteration (cache miss).
func BenchmarkMarshalDeflate(b *testing.B) {
	body := benchPayload(2000)
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		m := &Msg{Type: Publish, URI: "sport/events", body: body}
		sink, _ = m.MarshalDeflate()
	}
}
