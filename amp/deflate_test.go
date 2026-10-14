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

var syncMarker = []byte{0x00, 0x00, 0xff, 0xff}

// strictInflate decodes like a browser does. A sync-flushed payload has no final
// block, so io.ErrUnexpectedEOF is the expected outcome.
func strictInflate(payload []byte) ([]byte, error) {
	buf := bytes.NewBuffer(payload)
	buf.Write(syncMarker)
	r := flate.NewReader(buf)
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}

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

// Payloads escape to the caller; a later deflate must not write through them.
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

// Clients use server_no_context_takeover, so a pooled writer must not keep its window.
// Decoded without history, as a client would.
func TestDeflateDoesNotCarryDictionaryBetweenMessages(t *testing.T) {
	first := []byte(strings.Repeat("first-message-distinctive-content-", 500))
	second := []byte(strings.Repeat("second-message-entirely-different-", 500))

	// Prime the pooled writer.
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

// deflateWithDict compresses src with dict as the preceding window.
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

// The payload only decodes against a stale window, so Undeflate must fail on it.
func TestUndeflateDoesNotCarryDictionaryBetweenMessages(t *testing.T) {
	first := []byte(strings.Repeat("shared-window-content-", 400))
	second := append(bytes.Clone(first), []byte("-and-a-distinctive-tail")...)

	// Prime the pooled reader.
	if got := Undeflate(deflate(first)); !bytes.Equal(got, first) {
		t.Fatalf("priming decode failed: got %d bytes, want %d", len(got), len(first))
	}

	payload := deflateWithDict(t, second, first)

	// Sanity: the payload must need the dictionary.
	if got := inflateWithDict(t, payload, first); !bytes.Equal(got, second) {
		t.Fatalf("test payload does not decode even with its dictionary: got %d bytes, want %d",
			len(got), len(second))
	}

	if got := Undeflate(payload); bytes.Equal(got, second) {
		t.Fatal("the pooled decompressor resolved references into the previous message: " +
			"its window survived Reset, and every client decoding with an empty history would disagree")
	}
}

func TestUndeflateRecoversAfterMalformedPayload(t *testing.T) {
	src := []byte(strings.Repeat("valid-payload-", 500))
	good := deflate(src)

	_ = Undeflate([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	if got := Undeflate(good); !bytes.Equal(got, src) {
		t.Fatalf("a malformed payload poisoned the next decode: got %d bytes, want %d", len(got), len(src))
	}
}

// Deterministic version of the test above: sync.Pool may not return the primed reader.
func TestFlateResetClearsTheDecompressorWindow(t *testing.T) {
	first := []byte(strings.Repeat("shared-window-content-", 400))
	second := append(bytes.Clone(first), []byte("-and-a-distinctive-tail")...)
	payload := deflateWithDict(t, second, first)

	r := flate.NewReader(bytes.NewReader(nil)).(flateReader)
	defer func() { _ = r.Close() }()

	readAll := func() []byte {
		out, err := io.ReadAll(r)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			// Expected to fail without the dictionary.
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

	// Sanity: the payload must need the dictionary.
	resetTo(payload, first)
	if got := readAll(); !bytes.Equal(got, second) {
		t.Fatalf("payload does not decode even with its dictionary: got %d bytes, want %d", len(got), len(second))
	}

	// Populate the window.
	resetTo(deflate(first), nil)
	if got := readAll(); !bytes.Equal(got, first) {
		t.Fatalf("priming decode wrong: got %d bytes, want %d", len(got), len(first))
	}

	// Same reset as Undeflate.
	resetTo(payload, nil)
	if got := readAll(); bytes.Equal(got, second) {
		t.Fatal("Reset left the previous message in the decompressor window: " +
			"pooling a decompressor would leak one message into the next")
	}
}

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

// maxUndeflateSizeHint is an allocation budget, not an output limit.
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
