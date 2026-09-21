package amp

import (
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"
)

// syncMarker is the 4-byte tail that RFC 7692 section 7.2.1 requires a
// permessage-deflate sender to strip and the receiver to append back.
var syncMarker = []byte{0x00, 0x00, 0xff, 0xff}

// strictInflate is what a conforming peer (a browser, zlib) does with a
// permessage-deflate payload: append the sync marker and decode, reporting
// errors instead of swallowing them the way Undeflate does.
//
// It returns the decoded bytes and the terminating error. A sync-flushed
// message deliberately carries no final block -- the sender cut the stream at
// the marker and the peer resumes there -- so the decoder runs out of input at
// the end and io.ErrUnexpectedEOF is the expected, healthy outcome.
func strictInflate(payload []byte) ([]byte, error) {
	buf := bytes.NewBuffer(payload)
	buf.Write(syncMarker)
	r := flate.NewReader(buf)
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}

// deflateTestInputs covers both sides of compressionLenLimit and the shapes
// that decide how the encoder ends the stream.
func deflateTestInputs() []struct {
	name string
	data []byte
} {
	rnd := rand.New(rand.NewSource(1))
	random := make([]byte, 300000)
	rnd.Read(random)

	var bigJSON strings.Builder
	bigJSON.WriteString(`{"events":[`)
	for i := range 20000 {
		fmt.Fprintf(&bigJSON, `{"id":%d,"name":"event %d","odds":1.%03d,"type":"m"},`, i, i, i%1000)
	}
	bigJSON.WriteString(`{"end":0}],"rid":"_auto_:abcdef"}`)

	return []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"single byte", []byte("x")},
		{"tiny", []byte("hello world")},
		{"below compression limit", bytes.Repeat([]byte("ab"), 100)},
		{"at compression limit", bytes.Repeat([]byte("x"), compressionLenLimit)},
		{"just over compression limit", bytes.Repeat([]byte("y"), compressionLenLimit+1)},
		{"incompressible", random},
		{"big json snapshot", []byte(bigJSON.String())},
	}
}

// TestDeflateEndsAtSyncMarkerBoundary pins the contract that makes the 4-byte
// truncation exact instead of lucky: deflate must cut the stream exactly at the
// sync-flush marker, so that a peer re-appending 00 00 ff ff gets back the
// bytes the sender actually produced.
//
// Two things are asserted, and Close()-then-strip fails both:
//
//   - The peer must decode the payload cleanly and completely. Under Go 1.27
//     Close() emits a 2-byte fixed-Huffman final block, so stripping 4 bytes
//     eats 2 bytes of real data and the tail decodes to garbage.
//   - The decode must end by running out of input (io.ErrUnexpectedEOF), which
//     is the signature of a stream cut at a sync flush. Close() self-terminates
//     the stream with a final block instead, so the strip lands somewhere
//     inside it -- that is the latent defect, independent of toolchain.
func TestDeflateEndsAtSyncMarkerBoundary(t *testing.T) {
	for _, tc := range deflateTestInputs() {
		t.Run(tc.name, func(t *testing.T) {
			got, err := strictInflate(deflate(tc.data))
			if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("peer could not decode the payload: %v", err)
			}
			if !bytes.Equal(got, tc.data) {
				t.Fatalf("round trip mismatch: got %d bytes, want %d bytes\n got tail: %q\nwant tail: %q",
					len(got), len(tc.data), tail(got), tail(tc.data))
			}
			if err == nil {
				t.Fatalf("payload carries a final block: the stream was not cut at the sync marker, " +
					"so the 4 stripped bytes were not the marker")
			}
		})
	}
}

// TestDeflateRoundTripThroughUndeflate checks the pair svckit itself ships:
// deflate on the wire, Undeflate on the other end.
func TestDeflateRoundTripThroughUndeflate(t *testing.T) {
	for _, tc := range deflateTestInputs() {
		t.Run(tc.name, func(t *testing.T) {
			got := Undeflate(deflate(tc.data))
			if !bytes.Equal(got, tc.data) {
				t.Fatalf("round trip mismatch: got %d bytes, want %d bytes\n got tail: %q\nwant tail: %q",
					len(got), len(tc.data), tail(got), tail(tc.data))
			}
		})
	}
}

// TestMarshalDeflateRoundTrip exercises the path that actually broke in
// production: a Msg body over compressionLenLimit, marshaled for a client that
// negotiated permessage-deflate.
func TestMarshalDeflateRoundTrip(t *testing.T) {
	var body strings.Builder
	body.WriteString(`{"events":[`)
	for i := range 5000 {
		fmt.Fprintf(&body, `{"id":%d,"name":"event %d","type":"m"},`, i, i)
	}
	body.WriteString(`{"end":0}],"rid":"_auto_:2p1c8hv3btvif6"}`)

	m := &Msg{Type: Publish, URI: "sport/events", body: []byte(body.String())}

	payload, compressed := m.MarshalDeflate()
	if !compressed {
		t.Fatalf("body of %d bytes should have been compressed", body.Len())
	}

	plain := (&Msg{Type: Publish, URI: "sport/events", body: []byte(body.String())}).Marshal()
	got, err := strictInflate(payload)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("client could not decode the marshaled message: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("marshaled message did not survive the wire:\n got tail: %q\nwant tail: %q",
			tail(got), tail(plain))
	}
}

func tail(b []byte) string {
	const n = 32
	if len(b) > n {
		return "..." + string(b[len(b)-n:])
	}
	return string(b)
}

// TestUndeflateDoesNotWriteIntoCallerBuffer guards an aliasing hazard:
// bytes.NewBuffer adopts the slice it is given as its backing array, so
// appending the sync marker writes into the caller's spare capacity. On the
// receive path the payload is a view into a websocket read buffer, which is
// exactly the slice that must not be scribbled on.
func TestUndeflateDoesNotWriteIntoCallerBuffer(t *testing.T) {
	src := []byte(strings.Repeat("payload-", 2000))
	compressed := deflate(src)

	// A payload sitting in a larger read buffer, spare capacity behind it.
	const guard = 0xAA
	backing := bytes.Repeat([]byte{guard}, len(compressed)+64)
	copy(backing, compressed)
	payload := backing[:len(compressed)]

	if got := Undeflate(payload); !bytes.Equal(got, src) {
		t.Fatalf("round trip mismatch: got %d bytes, want %d", len(got), len(src))
	}

	for i, b := range backing[len(compressed):] {
		if b != guard {
			t.Fatalf("Undeflate wrote %#02x into the caller's buffer at +%d", b, i)
		}
	}
}
