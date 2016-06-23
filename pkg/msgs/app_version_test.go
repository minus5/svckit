package msgs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getAppVersion(t *testing.T, buf []byte) (*AppVersion) {
	av := &AppVersion{}
	err := json.Unmarshal(buf, av)
	assert.NoError(t, err)
	return av
}

func TestAppVersionStatKey(t *testing.T) {
	av := &AppVersion{}
	assert.Equal(t, "unknown.unknown", av.StatKey())

	av.App = "test"
	assert.Equal(t, "test.unknown", av.StatKey())

	av.Version = "1.0.0"
	assert.Equal(t, "test.1_0_0", av.StatKey())

	av = getAppVersion(t, []byte(`{"app":"test","version":"1.0.0"}`))
	assert.Equal(t, "test.1_0_0", av.StatKey())

	av = getAppVersion(t, []byte(`{"version":"1.0.0"}`))
	assert.Equal(t, "unknown.1_0_0", av.StatKey())

	av = getAppVersion(t, []byte(`{}`))
	assert.Equal(t, "unknown.unknown", av.StatKey())
}

// TODO: Test ToJson
// TODO: Test ToClient
// TODO: Test SameApp
// TODO: Test SameVersion
// TODO: Test SetExpiresFor


