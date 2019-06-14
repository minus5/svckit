package msgs

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
)

//AppVersion verzija za neki tip aplikacije.
type AppVersion struct {
	// Naziv aplikacije
	App				string
	// Verzija aplikacije
	Version			string
	// Stranica na kojoj se aplikacija nalazi prilikom slanja poruke, primamo samo radi raspisa statistike
	Page			string	`json:"page,omitempty"`
	// postotak klijenata za koje spremamo logove i saljemo statistike
	DebugPostotak	int		`bson:"debug_postotak" json:"debug_postotak"`
	// unix timestamp do kad je verziju aplikacije moguce koristiti 
	ExpiresAt		int		`json:"expires_at"`
	// array verzija aplikacije s njihovim datumom do kad ih je moguce korisiti
	Valid			[]struct {
		// Verzija aplikacije
		Version		string
		// unix timestamp do kad je verziju aplikacije moguce koristiti 
		ExpiresAt	int		`bson:"expires_at" json:"expires_at"`
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

// Vraca string za statistike u formatu app.version, unknown se koristi za vrijednosti koje su prazne
// verzija umjesto tocke sadrzi underscore npr 1.1.1 -> 1_1_1
func (av *AppVersion) StatKey() string {
	app := "unknown"
	ver := "unknown"
	if av.App != "" {
		app = av.App
	}
	if av.Version != "" {
		ver = strings.Replace(av.Version, ".", "_", -1)
	}
	return fmt.Sprintf("%s.%s", app, ver)
}

// Da li je jednaka aplikacija?
func (av *AppVersion) SameApp(other *AppVersion) bool {
	return av.App == other.App
}

// Da li je jednaka verzija aplikacije?
func (av *AppVersion) SameVersion(other *AppVersion) bool {
	return av.Version == other.Version
}

// Podaci vezije aplikacija koji se salju klijentima
func (av *AppVersion) ToClient(uvijekDebug bool) []byte {
	d := struct {
		App       string `json:"app"`
		Version   string `json:"version"`
		ExpiresAt int    `json:"expires_at,omitempty"`
		Debug     bool   `json:"debug"`
	}{
		App:       av.App,
		Version:   av.Version,
		ExpiresAt: av.ExpiresAt,
		// flag da li klijent aplikacija smije slati logove i metrike
		Debug:     rand.Intn(100) <= av.DebugPostotak || uvijekDebug,
	}
	buf, _ := json.Marshal(d)
	return buf
}
