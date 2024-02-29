package amp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURI(t *testing.T) {
	topic := "hr.mnu5"
	path := "resource/method"
	m := NewPublish(topic, path, 123, Full, nil)

	assert.Equal(t, path, m.Path())
	assert.Equal(t, topic, m.Topic())
	assert.Equal(t, topic+"/"+path, m.URI)
}

func TestURIWithoutPath(t *testing.T) {
	topic := "hr.mnu5"
	path := ""
	m := NewPublish(topic, path, 123, Full, nil)

	assert.Equal(t, path, m.Path())
	assert.Equal(t, topic, m.Topic())
	assert.Equal(t, topic, m.URI)
}

func TestPublish(t *testing.T) {
	o := struct {
		First string
		Last  string
	}{First: "jozo", Last: "bozo"}
	topic := "hr.mnu5"
	path := "resource/method"
	m := NewPublish(topic, path, 123, Full, o)

	buf := m.Marshal()
	expected := `{"u":"hr.mnu5/resource/method","s":123,"p":1}
{"First":"jozo","Last":"bozo"}`

	assert.Equal(t, expected, string(buf))
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want *Msg
	}{
		{
			name: "it should return nil if input is nil",
			in:   nil,
			want: nil,
		},
		{
			name: "it should parse the message, ignoring values in 'h' field",
			in:   []byte(`{"t":2,"u":"some.topic/method","i":4,"b":{"topic.one":1},"m":{"a":"b"},"h":{"a":"b"}}`),
			want: &Msg{
				Type:          Request,
				URI:           "some.topic/method",
				CorrelationID: 4,
				Subscriptions: map[string]int64{
					"topic.one": 1,
				},
				Meta: map[string]string{
					"a": "b",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseFromBackend(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want *Msg
	}{
		{
			name: "it should return nil if input is nil",
			in:   nil,
			want: nil,
		},
		{
			name: "it should parse the message",
			in:   []byte(`{"t":2,"u":"some.topic/method","i":4,"h":{"a":"b"}}`),
			want: &Msg{
				Type:          Request,
				URI:           "some.topic/method",
				CorrelationID: 4,
				BackendHeaders: map[string]string{
					"a": "b",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFromBackend(tt.in)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseV1Subscribe(t *testing.T) {
	buf := `{"t":1,"u":[{"s":"m","n":93601933},{"s":"d_174626231","n":10},{"s":"s_2","n":11},{"s":"s_4","n":12}]}`
	m := ParseV1([]byte(buf))
	assert.NotNil(t, m)
	assert.Len(t, m.Subscriptions, 4)
	assert.Equal(t, m.Subscriptions["sportsbook/m"], int64(93601933))
	assert.Equal(t, m.Subscriptions["sportsbook/s_2"], int64(11))
	assert.Equal(t, m.Subscriptions["sportsbook/s_4"], int64(12))
	assert.Equal(t, m.Subscriptions["sportsbook/d_174626231"], int64(10))
}

func TestParseV1Ping(t *testing.T) {
	buf := `{"t":4}`
	m := ParseV1([]byte(buf))
	assert.NotNil(t, m)
	assert.Equal(t, m.Type, Ping)
}

func TestMarshalV1(t *testing.T) {
	m := &Msg{
		Type:       Publish,
		URI:        "sportsbook/m",
		Ts:         123,
		UpdateType: Full,
		body:       []byte(`{"First":"jozo","Last":"bozo"}`),
	}
	buf := m.MarshalV1()
	assert.Equal(t, string(buf), `{"s":"m","n":123,"f":1}
{"First":"jozo","Last":"bozo"}`)

	m.payloads = nil
	m.UpdateType = Diff
	buf = m.MarshalV1()
	assert.Equal(t, string(buf), `{"s":"m","n":123}
{"First":"jozo","Last":"bozo"}`)

	buf = m.Marshal()
	assert.Equal(t, string(buf), `{"u":"sportsbook/m","s":123}
{"First":"jozo","Last":"bozo"}`)

}

func TestParseV1Subscriptions(t *testing.T) {
	buf := []byte(`[{"s":"m","n":94067395},{"s":"s_4","n":1},{"s":"s_5","n":2}]`)
	m := ParseV1Subscriptions(buf)
	assert.NotNil(t, m)

	assert.Len(t, m.Subscriptions, 3)
	assert.Equal(t, m.Subscriptions["sportsbook/m"], int64(94067395))
	assert.Equal(t, m.Subscriptions["sportsbook/s_4"], int64(1))
	assert.Equal(t, m.Subscriptions["sportsbook/s_5"], int64(2))
}

func TestMsg_Marshal(t *testing.T) {
	tests := []struct {
		name string
		in   *Msg
		want []byte
	}{
		{
			name: "it should marshal the message without body",
			in: &Msg{
				Type:          Response,
				CorrelationID: 4,
				Error: &Error{
					Message: "error",
				},
				URI:        "some.topic",
				Ts:         12,
				UpdateType: Full,
				Replay:     Replay,
				CacheDepth: 5,
				Meta: map[string]string{
					"a": "b",
				},
			},
			want: append(
				[]byte(`{"t":3,"i":4,"e":{"m":"error"},"u":"some.topic","s":12,"p":1,"l":1,"d":5,"m":{"a":"b"}}`),
				[]byte("\n")...,
			),
		},
		{
			name: "it should marshal the message",
			in: &Msg{
				Type:          Response,
				CorrelationID: 4,
				Error: &Error{
					Message: "error",
				},
				URI:        "some.topic",
				Ts:         12,
				UpdateType: Full,
				Replay:     Replay,
				CacheDepth: 5,
				Meta: map[string]string{
					"a": "b",
				},
				body: []byte(`{"foo":"bar"}`),
			},
			want: append(
				append(
					[]byte(`{"t":3,"i":4,"e":{"m":"error"},"u":"some.topic","s":12,"p":1,"l":1,"d":5,"m":{"a":"b"}}`),
					[]byte("\n")...,
				),
				[]byte(`{"foo":"bar"}`)...,
			),
		},
		{
			name: "it should marshal the message that contains BackendHeaders without adding that headers to the marshalled payload",
			in: &Msg{
				Type:          Response,
				CorrelationID: 4,
				Error: &Error{
					Message: "error",
				},
				URI:        "some.topic",
				Ts:         12,
				UpdateType: Full,
				Replay:     Replay,
				CacheDepth: 5,
				Meta: map[string]string{
					"a": "b",
				},
				BackendHeaders: map[string]string{
					"foo": "bar",
					"bar": "baz",
				},
				body: []byte(`{"foo":"bar"}`),
			},
			want: append(
				append(
					[]byte(`{"t":3,"i":4,"e":{"m":"error"},"u":"some.topic","s":12,"p":1,"l":1,"d":5,"m":{"a":"b"}}`),
					[]byte("\n")...,
				),
				[]byte(`{"foo":"bar"}`)...,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Marshal()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestMsg_MarshalForBackend(t *testing.T) {
	tests := []struct {
		name string
		in   *Msg
		want []byte
	}{
		{
			name: "it should marshal the message without body",
			in: &Msg{
				Type:          Response,
				CorrelationID: 4,
				Meta: map[string]string{
					"a": "b",
				},
			},
			want: append(
				[]byte(`{"t":3,"i":4,"m":{"a":"b"}}`),
				[]byte("\n")...,
			),
		},
		{
			name: "it should marshal the message",
			in: &Msg{
				Type:          Response,
				CorrelationID: 4,
				Meta: map[string]string{
					"a": "b",
				},
				body: []byte(`{"foo":"bar"}`),
			},
			want: append(
				append(
					[]byte(`{"t":3,"i":4,"m":{"a":"b"}}`),
					[]byte("\n")...,
				),

				[]byte(`{"foo":"bar"}`)...,
			),
		},
		{
			name: "it should marshal the message that contains BackendHeaders, adding that headers to the marshalled payload",
			in: &Msg{
				Type:          Response,
				CorrelationID: 4,
				Meta: map[string]string{
					"a": "b",
				},
				BackendHeaders: map[string]string{
					"c": "d",
				},
				body: []byte(`{"foo":"bar"}`),
			},
			want: append(
				append(
					[]byte(`{"t":3,"i":4,"m":{"a":"b"},"h":{"c":"d"}}`),
					[]byte("\n")...,
				),

				[]byte(`{"foo":"bar"}`)...,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.MarshalForBackend()
			require.Equal(t, tt.want, got)
		})
	}
}
