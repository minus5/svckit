package statsd

import (
	"net"
	"strings"
	"testing"
	"time"

	api "github.com/smira/go-statsd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	incrStat    string
	incrCount   int64
	incrTags    []api.Tag
	gaugeStat   string
	gaugeValue  int64
	gaugeTags   []api.Tag
	timingStat  string
	timingDelta int64
	timingTags  []api.Tag
}

func (f *fakeClient) Incr(stat string, count int64, tags ...api.Tag) {
	f.incrStat, f.incrCount, f.incrTags = stat, count, tags
}
func (f *fakeClient) Decr(string, int64, ...api.Tag) {}
func (f *fakeClient) Timing(stat string, delta int64, tags ...api.Tag) {
	f.timingStat, f.timingDelta, f.timingTags = stat, delta, tags
}
func (f *fakeClient) PrecisionTiming(string, time.Duration, ...api.Tag) {}
func (f *fakeClient) Gauge(stat string, value int64, tags ...api.Tag) {
	f.gaugeStat, f.gaugeValue, f.gaugeTags = stat, value, tags
}
func (f *fakeClient) GaugeDelta(string, int64, ...api.Tag)    {}
func (f *fakeClient) FGauge(string, float64, ...api.Tag)      {}
func (f *fakeClient) FGaugeDelta(string, float64, ...api.Tag) {}
func (f *fakeClient) SetAdd(string, string, ...api.Tag)       {}
func (f *fakeClient) Close() error                            { return nil }

// The label variants must apply the same prefix as their unlabelled
// counterparts and convert the labels map to tags sorted by name, so the
// emitted line does not depend on Go map iteration order.
func TestLabelVariantsForwardSortedTags(t *testing.T) {
	f := &fakeClient{}
	s := newStatsd("app.", f)
	labels := map[string]string{"version": "2026.0820.094120", "terminal_name": "SSBT 5007_1"}
	wantTags := []api.Tag{
		api.StringTag("terminal_name", "SSBT 5007_1"),
		api.StringTag("version", "2026.0820.094120"),
	}

	s.GaugeL("frontend_version", labels, 1787564529)
	assert.Equal(t, "app.frontend_version", f.gaugeStat)
	assert.Equal(t, int64(1787564529), f.gaugeValue)
	assert.Equal(t, wantTags, f.gaugeTags)

	s.TimeL("frontend_time", labels, 42)
	assert.Equal(t, "app.frontend_time", f.timingStat)
	assert.Equal(t, int64(42), f.timingDelta)
	assert.Equal(t, wantTags, f.timingTags)
}

// CounterL keeps Counter's value semantics: no values means one increment,
// values mean their sum.
func TestCounterLValueSemantics(t *testing.T) {
	f := &fakeClient{}
	s := newStatsd("app.", f)

	s.CounterL("hits", map[string]string{"k": "v"})
	assert.Equal(t, "app.hits", f.incrStat)
	assert.Equal(t, int64(1), f.incrCount)
	assert.Equal(t, []api.Tag{api.StringTag("k", "v")}, f.incrTags)

	s.CounterL("hits", map[string]string{"k": "v"}, 2, 3)
	assert.Equal(t, int64(5), f.incrCount)
}

// Empty and nil label maps must degrade to the plain untagged call, so a
// label variant is always safe to use.
func TestLabelVariantsWithoutLabels(t *testing.T) {
	f := &fakeClient{}
	s := newStatsd("app.", f)

	s.GaugeL("g", nil, 7)
	assert.Equal(t, "app.g", f.gaugeStat)
	assert.Empty(t, f.gaugeTags)

	s.GaugeL("g", map[string]string{}, 7)
	assert.Empty(t, f.gaugeTags)
}

// The exact wire line, through the real go-statsd client with the DogStatsD
// tag style Dial configures: tags trail the type as `|#k:v,k:v`, and a space
// inside a tag value passes through untouched. statsd_exporter parses this
// dialect by default and exposes each tag as a Prometheus label. The InfluxDB
// style was rejected on purpose: its tags sit inside the name section, so a
// raw-statsd relay to graphite would mint one graphite metric name per tag
// combination.
func TestGaugeLWireFormat(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer conn.Close()

	c := api.NewClient(conn.LocalAddr().String(),
		api.TagStyle(api.TagFormatDatadog),
		api.FlushInterval(time.Millisecond),
	)
	s := newStatsd("retail_pl.", c)

	s.GaugeL("ssbt_frontend_version", map[string]string{
		"version":       "2026.0820.094120",
		"terminal_name": "SSBT 5007_1",
	}, 1787564529)
	require.NoError(t, c.Close())

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
	buf := make([]byte, 1024)
	n, _, err := conn.ReadFrom(buf)
	require.NoError(t, err)
	assert.Equal(t,
		"retail_pl.ssbt_frontend_version:1787564529|g|#terminal_name:SSBT 5007_1,version:2026.0820.094120",
		strings.TrimSpace(string(buf[:n])))
}
