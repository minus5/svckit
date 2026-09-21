package amp

import (
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
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

// TestDeflateDoesNotRecyclePayload is the guard on pooling: only the codec
// state may be reused between calls. The returned payload is cached in
// Msg.payloads and handed to a websocket writer, so a later deflate must never
// write through it.
func TestDeflateDoesNotRecyclePayload(t *testing.T) {
	first := deflate([]byte(strings.Repeat("first-", 3000)))
	snapshot := bytes.Clone(first)

	for range 50 {
		_ = deflate([]byte(strings.Repeat("second-", 3000)))
	}

	if !bytes.Equal(first, snapshot) {
		t.Fatal("a later deflate wrote through a payload already returned to a caller")
	}
}

// TestDeflateConcurrent drives the compress and decompress paths from many
// goroutines with a distinct payload each. Pooled codec state leaking across
// goroutines would surface as a round-trip mismatch (and -race would flag it).
func TestDeflateConcurrent(t *testing.T) {
	const goroutines, iters = 64, 50
	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Go(func() {
			payload := []byte(strings.Repeat(fmt.Sprintf("goroutine-%d-payload-", g), 500))
			for range iters {
				got := Undeflate(deflate(payload))
				if !bytes.Equal(got, payload) {
					t.Errorf("g%d: round-trip mismatch (got %d bytes, want %d)", g, len(got), len(payload))
					return
				}
			}
		})
	}
	wg.Wait()
}

// TestDeflateDoesNotCarryDictionaryBetweenMessages is the test that pooling
// most needs and that a naive round-trip test cannot give.
//
// The handshake negotiates server_no_context_takeover (amp/ws/listener.go), so
// the client decompresses every message with an empty history. A pooled writer
// that kept its LZ77 window would emit back-references into the previous
// message, which the client cannot resolve.
//
// Decoding here therefore has to use a decompressor with no history of its own:
// a pooled reader would carry the same leaked window and the round trip would
// agree with itself while every real client saw garbage.
func TestDeflateDoesNotCarryDictionaryBetweenMessages(t *testing.T) {
	first := []byte(strings.Repeat("first-message-distinctive-content-", 500))
	second := []byte(strings.Repeat("second-message-entirely-different-", 500))

	// Prime the pooled writer several times so a later call is guaranteed to get
	// a writer that has already compressed something else.
	for range 5 {
		_ = deflate(first)
	}

	got, err := strictInflate(deflate(second))
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("a client with no history could not decode the message: %v", err)
	}
	if !bytes.Equal(got, second) {
		t.Fatalf("message decoded against an empty history is wrong: got %d bytes, want %d\n got tail: %q",
			len(got), len(second), tail(got))
	}
}

// deflateWithDict compresses src the way a sender with context takeover would:
// the previous message stays in the LZ77 window, so the output may reference it
// and only a decompressor holding that same window can resolve it.
func deflateWithDict(t *testing.T, src, dict []byte) []byte {
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
	return buf[:len(buf)-len(syncMarker)]
}

// inflateWithDict decodes a payload against a preset window.
func inflateWithDict(t *testing.T, payload, dict []byte) []byte {
	t.Helper()
	buf := bytes.NewBuffer(bytes.Clone(payload))
	buf.Write(syncMarker)
	r := flate.NewReaderDict(buf, dict)
	defer func() { _ = r.Close() }()
	out, err := io.ReadAll(r)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("decode with dictionary: %v", err)
	}
	return out
}

// TestUndeflateDoesNotCarryDictionaryBetweenMessages is the mirror image on the
// receive side. It cannot be written as an ordinary round trip: a correctly
// compressed message is self-contained, so a decompressor that wrongly kept its
// previous window decodes it just as well as one that reset, and the test would
// pass either way.
//
// The payload therefore has to be one that only a stale window can resolve --
// compressed with the previous message as its dictionary, the way a sender with
// context takeover would emit it. Undeflate must fail to reconstruct it. That is
// the assertion: the pooled decompressor carries nothing from the message
// before, so references into it cannot resolve.
//
// This covers the real path but cannot be the only guard, because sync.Pool does
// not promise to return the instance primed above. TestFlateResetClearsThe-
// DecompressorWindow holds the reader itself and settles the same property
// deterministically.
func TestUndeflateDoesNotCarryDictionaryBetweenMessages(t *testing.T) {
	first := []byte(strings.Repeat("shared-window-content-", 400))
	second := append(bytes.Clone(first), []byte("-and-a-distinctive-tail")...)

	// Prime the pooled reader: its window would hold first's output if Reset
	// failed to clear it.
	if got := Undeflate(deflate(first)); !bytes.Equal(got, first) {
		t.Fatalf("priming decode failed: got %d bytes, want %d", len(got), len(first))
	}

	payload := deflateWithDict(t, second, first)

	// The payload must genuinely depend on the dictionary, or the assertion
	// below would hold for the wrong reason.
	if got := inflateWithDict(t, payload, first); !bytes.Equal(got, second) {
		t.Fatalf("test payload does not decode even with its dictionary: got %d bytes, want %d",
			len(got), len(second))
	}

	if got := Undeflate(payload); bytes.Equal(got, second) {
		t.Fatal("the pooled decompressor resolved references into the previous message: " +
			"its window survived Reset, and every client decoding with an empty history would disagree")
	}
}

// TestUndeflateRecoversAfterMalformedPayload guards the reader pool in this
// package, which is separate from the one in amp/ws: a payload that fails to
// decode leaves the reader in an error state, and reusing it without a full
// reset would break the next, valid message.
func TestUndeflateRecoversAfterMalformedPayload(t *testing.T) {
	src := []byte(strings.Repeat("valid-payload-", 500))
	good := deflate(src)

	_ = Undeflate([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	if got := Undeflate(good); !bytes.Equal(got, src) {
		t.Fatalf("a malformed payload poisoned the next decode: got %d bytes, want %d", len(got), len(src))
	}
}

// TestFlateResetClearsTheDecompressorWindow pins the standard library guarantee
// the whole reader pool rests on: Reset with a nil dictionary starts the
// decompressor from an empty window, so a pooled instance cannot leak one
// message into the next.
//
// The behavioral tests reach this through Undeflate, which leaves them at the
// mercy of sync.Pool handing back the instance they primed -- it promises no
// such thing, and a pool that dropped the instance would supply a fresh reader
// with no window either, so a broken reset could pass unnoticed. This test holds
// the reader itself and cannot miss it.
func TestFlateResetClearsTheDecompressorWindow(t *testing.T) {
	first := []byte(strings.Repeat("shared-window-content-", 400))
	second := append(bytes.Clone(first), []byte("-and-a-distinctive-tail")...)
	payload := deflateWithDict(t, second, first)

	r := flate.NewReader(bytes.NewReader(nil)).(flateReader)
	defer func() { _ = r.Close() }()

	readAll := func() []byte {
		out, err := io.ReadAll(r)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			// Decoding a dictionary-dependent payload without the dictionary is
			// expected to end badly; anything else is a broken test.
			var corrupt flate.CorruptInputError
			if !errors.As(err, &corrupt) {
				t.Fatalf("decode: %v", err)
			}
		}
		return out
	}
	resetTo := func(data []byte, dict []byte) {
		if err := r.Reset(io.MultiReader(bytes.NewReader(data), bytes.NewReader(syncMarker)), dict); err != nil {
			t.Fatalf("reset: %v", err)
		}
	}

	// The payload must genuinely need the previous message as its window, or
	// the assertion below would hold for the wrong reason.
	resetTo(payload, first)
	if got := readAll(); !bytes.Equal(got, second) {
		t.Fatalf("payload does not decode even with its dictionary: got %d bytes, want %d", len(got), len(second))
	}

	// Give the reader a window the ordinary way, by decoding a message.
	resetTo(deflate(first), nil)
	if got := readAll(); !bytes.Equal(got, first) {
		t.Fatalf("priming decode wrong: got %d bytes, want %d", len(got), len(first))
	}

	// This is the reset Undeflate performs before every message.
	resetTo(payload, nil)
	if got := readAll(); bytes.Equal(got, second) {
		t.Fatal("Reset left the previous message in the decompressor window: " +
			"pooling a decompressor would leak one message into the next")
	}
}

// TestUndeflatedSizeHintStaysInRange guards the arithmetic in the size hint.
// Undeflate is exported and its argument is caller-supplied, so a length large
// enough to overflow compressed*8 would produce a negative capacity and panic
// in make before any decoding happened.
func TestUndeflatedSizeHintStaysInRange(t *testing.T) {
	for _, compressed := range []int{
		0, 1, 1024,
		maxUndeflateSizeHint/8 - 1, maxUndeflateSizeHint / 8, maxUndeflateSizeHint/8 + 1,
		1 << 30, math.MaxInt / 4, math.MaxInt,
	} {
		got := undeflatedSizeHint(compressed)
		if got < 0 || got > maxUndeflateSizeHint {
			t.Errorf("undeflatedSizeHint(%d) = %d, want within [0, %d]", compressed, got, maxUndeflateSizeHint)
		}
	}
}

// TestUndeflateDecodesBeyondTheSizeHint pins that maxUndeflateSizeHint is a
// budget for the first allocation and not a limit on the message. Both branches
// of the hint have to survive output larger than the cap: the repetitive case
// asks for less than the cap and grows past it, the incompressible case has its
// request clamped to the cap and grows past it too.
func TestUndeflateDecodesBeyondTheSizeHint(t *testing.T) {
	rnd := rand.New(rand.NewSource(2))
	incompressible := make([]byte, maxUndeflateSizeHint*3/2)
	rnd.Read(incompressible)

	for _, tc := range []struct {
		name      string
		data      []byte
		wantedCap bool // whether the hint gets clamped to maxUndeflateSizeHint
	}{
		{"repetitive", bytes.Repeat([]byte("repeated-content-"), maxUndeflateSizeHint*3/2/17), false},
		{"incompressible", incompressible, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.data) <= maxUndeflateSizeHint {
				t.Fatalf("payload of %d bytes does not exceed the %d byte hint", len(tc.data), maxUndeflateSizeHint)
			}
			compressed := deflate(tc.data)
			if clamped := undeflatedSizeHint(len(compressed)) == maxUndeflateSizeHint; clamped != tc.wantedCap {
				t.Fatalf("hint clamped = %v, want %v (compressed to %d bytes)", clamped, tc.wantedCap, len(compressed))
			}

			if got := Undeflate(compressed); !bytes.Equal(got, tc.data) {
				t.Fatalf("output was truncated: got %d bytes, want %d", len(got), len(tc.data))
			}
		})
	}
}
