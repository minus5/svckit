package msgs

import (
	"encoding/json"
	"fmt"
	"strings"
)

//AppVersion verzija za neki tip aplikacije.
type AppVersion struct {
	App       string
	Version   string
	Debug     bool `json:"debug"`
	ExpiresAt int  `json:"expires_at"`
	Valid     []struct {
		Version   string
		ExpiresAt int `bson:"expires_at" json:"expires_at"`
	}
}

//ToJson serijalizira u json.
func (av *AppVersion) ToJson() []byte {
	buf, _ := json.Marshal(av)
	return buf
}

//SetExpiresFor za neku client verziju nadje kada je expires
func (av *AppVersion) SetExpiresFor(cv string) {
	if av.Version == cv {
		av.ExpiresAt = 0
		return
	}
	for _, v := range av.Valid {
		if v.Version == cv {
			av.ExpiresAt = v.ExpiresAt
			return
		}
	}
	av.ExpiresAt = -1
}

func (av *AppVersion) StatKey() string {
	if av.App == "" && av.Version == "" {
		return "unknown"
	}
	return fmt.Sprintf("%s.%s", av.App, strings.Replace(av.Version, ".", "_", -1))
}

func (av *AppVersion) SameApp(other *AppVersion) bool {
	return av.App == other.App
}

func (av *AppVersion) SameVersion(other *AppVersion) bool {
	return av.Version == other.Version
}

func (av *AppVersion) ToClient() []byte {
	d := struct {
		App       string `json:"app"`
		Version   string `json:"version"`
		ExpiresAt int    `json:"expires_at,omitempty"`
		Debug     bool   `json:"debug"`
	}{
		App:       av.App,
		Version:   av.Version,
		ExpiresAt: av.ExpiresAt,
		Debug:     av.Debug,
	}
	buf, _ := json.Marshal(d)
	return buf
}
