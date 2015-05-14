package jsonreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	t.Skip()
	jr := New("http://localhost:8091/tecajna/zadrske")
	rsp, err := jr.Get()
	assert.Nil(t, err)
	assert.Equal(t, 200, jr.StatusCode())
	assert.Equal(t, string(rsp),
		`[{"sportId":2,"liga":"Poljska - Ekstraklasa","zadrska":10},{"sportId":2,"liga":"Rumunjska - Kup","zadrska":12}]`)
	assert.Equal(t, "260527", jr.VersionHeader())

	jr = New("http://localhost:8091/tecajna/zadrske",
		VersionHeader("260526"))
	rsp, err = jr.Post(nil)
	assert.Nil(t, err)
	assert.Equal(t, 200, jr.StatusCode())
	assert.Equal(t, string(rsp),
		`[{"sportId":2,"liga":"Poljska - Ekstraklasa","zadrska":10}]`)
	assert.Equal(t, "260527", jr.VersionHeader())

	//New("http://localhost:8090/tecajna/zadrske").Get()
}

func TestIntegrationGetWithVersion(t *testing.T) {
	t.Skip()
	v := ""
	rsp, v2, err := GetWithVersion("http://localhost:8091/tecajna/zadrske", v)
	assert.Nil(t, err)
	assert.Equal(t, string(rsp),
		`[{"sportId":2,"liga":"Poljska - Ekstraklasa","zadrska":10},{"sportId":2,"liga":"Rumunjska - Kup","zadrska":12}]`)
	assert.Equal(t, "260527", v2)

	rsp, v3, err := GetWithVersion("http://localhost:8091/tecajna/zadrske", "260526")
	assert.Nil(t, err)
	assert.Equal(t, string(rsp),
		`[{"sportId":2,"liga":"Poljska - Ekstraklasa","zadrska":10}]`)
	assert.Equal(t, "260527", v3)

	//New("http://localhost:8090/tecajna/zadrske").Get()
}

func TestCalcRetryInterval(t *testing.T) {
	r := New("")
	r.retrySleep = 1000
	assert.Equal(t, 1, r.calcRetryInterval(0))
	assert.Equal(t, 2, r.calcRetryInterval(1))
	assert.Equal(t, 7, r.calcRetryInterval(2))
	assert.Equal(t, 20, r.calcRetryInterval(3))
	assert.Equal(t, 54, r.calcRetryInterval(4))
	assert.Equal(t, 148, r.calcRetryInterval(5))
	assert.Equal(t, 403, r.calcRetryInterval(6))
	assert.Equal(t, 1000, r.calcRetryInterval(7))
}
