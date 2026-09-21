package ws

import (
	"bytes"
	"compress/flate"
	"strings"
	"testing"
)

// wireDeflate produces a permessage-deflate payload the way a conforming peer
// does: sync-flush the stream and cut the trailing 00 00 ff ff marker
// (RFC 7692 section 7.2.1).
func wireDeflate(t *testing.T, src []byte) []byte {
	t.Helper()
	var dest bytes.Buffer
	w, err := flate.NewWriter(&dest, flate.DefaultCompression)
	if err != nil {
		t.Fatalf("flate.NewWriter: %v", err)
	}
	if _, err := w.Write(src); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	buf := dest.Bytes()
	if !bytes.HasSuffix(buf, []byte{0x00, 0x00, 0xff, 0xff}) {
		t.Fatalf("sync flush did not end with the marker: % x", buf[max(0, len(buf)-4):])
	}
	return buf[:len(buf)-4]
}

func TestUndeflateRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"tiny", []byte("hello world")},
		{"repetitive", bytes.Repeat([]byte("subscribe sport/events "), 500)},
		{"large", []byte(strings.Repeat(`{"id":1,"type":"m"},`, 50000))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := undeflate(wireDeflate(t, tc.data)); !bytes.Equal(got, tc.data) {
				t.Fatalf("round trip mismatch: got %d bytes, want %d", len(got), len(tc.data))
			}
		})
	}
}

// TestUndeflateDoesNotWriteIntoCallerBuffer guards an aliasing hazard:
// bytes.NewBuffer adopts the slice it is given as its backing array, so
// appending the sync marker writes into the caller's spare capacity. Here the
// payload is a view into the websocket read buffer, which must not be touched.
func TestUndeflateDoesNotWriteIntoCallerBuffer(t *testing.T) {
	src := []byte(strings.Repeat("payload-", 2000))
	compressed := wireDeflate(t, src)

	const guard = 0xAA
	backing := bytes.Repeat([]byte{guard}, len(compressed)+64)
	copy(backing, compressed)
	payload := backing[:len(compressed)]

	if got := undeflate(payload); !bytes.Equal(got, src) {
		t.Fatalf("round trip mismatch: got %d bytes, want %d", len(got), len(src))
	}
	for i, b := range backing[len(compressed):] {
		if b != guard {
			t.Fatalf("undeflate wrote %#02x into the caller's buffer at +%d", b, i)
		}
	}
}
