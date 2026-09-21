package ws

import (
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
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

// TestUndeflateRecoversAfterMalformedPayload guards the reader pool: a payload
// that fails to decode leaves the reader in an error state, and reusing it
// without a full reset would break the next, valid message.
func TestUndeflateRecoversAfterMalformedPayload(t *testing.T) {
	src := []byte(strings.Repeat("valid-payload-", 500))
	good := wireDeflate(t, src)

	_ = undeflate([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	if got := undeflate(good); !bytes.Equal(got, src) {
		t.Fatalf("a malformed payload poisoned the next decode: got %d bytes, want %d", len(got), len(src))
	}
}

// TestUndeflateConcurrent drives the pooled decompressor from many goroutines
// with a distinct payload each; pooled state leaking across goroutines would
// surface as a mismatch, and -race would flag a data race.
func TestUndeflateConcurrent(t *testing.T) {
	const goroutines, iters = 64, 50
	payloads := make([][]byte, goroutines)
	sources := make([][]byte, goroutines)
	for g := range goroutines {
		sources[g] = []byte(strings.Repeat(fmt.Sprintf("goroutine-%d-payload-", g), 500))
		payloads[g] = wireDeflate(t, sources[g])
	}

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			for range iters {
				if got := undeflate(payloads[g]); !bytes.Equal(got, sources[g]) {
					t.Errorf("g%d: round-trip mismatch (got %d bytes, want %d)", g, len(got), len(sources[g]))
					return
				}
			}
		})
	}
	wg.Wait()
}

// wireDeflateWithDict compresses src the way a sender with context takeover
// would: the previous message stays in the LZ77 window, so the output may
// reference it and only a decompressor holding that same window can resolve it.
func wireDeflateWithDict(t *testing.T, src, dict []byte) []byte {
	t.Helper()
	var dest bytes.Buffer
	w, err := flate.NewWriterDict(&dest, flate.DefaultCompression, dict)
	if err != nil {
		t.Fatalf("flate.NewWriterDict: %v", err)
	}
	if _, err := w.Write(src); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	buf := dest.Bytes()
	return buf[:len(buf)-4]
}

// TestUndeflateDoesNotCarryDictionaryBetweenMessages checks that the pooled
// decompressor starts every message with an empty window.
//
// It cannot be written as an ordinary round trip: a correctly compressed message
// is self-contained, so a decompressor that wrongly kept its previous window
// decodes it just as well as one that reset, and the test would pass either way.
// The payload is therefore one that only a stale window can resolve, and
// undeflate must fail to reconstruct it.
//
// The connection negotiates client_no_context_takeover (listener.go), so a client
// that ignored that and compressed against its history must not be silently
// understood here either.
//
// This exercises the real path but rests on sync.Pool returning the instance
// primed above, which it does not promise. The reset contract itself is settled
// deterministically by TestFlateResetClearsTheDecompressorWindow in amp, and the
// pool here is reset the same way.
func TestUndeflateDoesNotCarryDictionaryBetweenMessages(t *testing.T) {
	first := []byte(strings.Repeat("shared-window-content-", 400))
	second := append(bytes.Clone(first), []byte("-and-a-distinctive-tail")...)

	if got := undeflate(wireDeflate(t, first)); !bytes.Equal(got, first) {
		t.Fatalf("priming decode failed: got %d bytes, want %d", len(got), len(first))
	}

	payload := wireDeflateWithDict(t, second, first)

	// The payload must genuinely depend on the dictionary, or the assertion
	// below would hold for the wrong reason.
	ref := bytes.NewBuffer(bytes.Clone(payload))
	ref.Write([]byte{0x00, 0x00, 0xff, 0xff})
	r := flate.NewReaderDict(ref, first)
	withDict, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("decode with dictionary: %v", err)
	}
	if !bytes.Equal(withDict, second) {
		t.Fatalf("test payload does not decode even with its dictionary: got %d bytes, want %d",
			len(withDict), len(second))
	}

	if got := undeflate(payload); bytes.Equal(got, second) {
		t.Fatal("the pooled decompressor resolved references into the previous message: " +
			"its window survived Reset")
	}
}
