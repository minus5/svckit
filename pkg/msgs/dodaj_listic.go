package msgs

import "encoding/json"

type DodajListic struct {
	Broj          string `json:"broj"`
	KontrolniBroj string `json:"kontrolni_broj"`
	IgracId       string `json:"igrac_guid"`
}

func (dl *DodajListic) ToJson() []byte {
	buf, _ := json.Marshal(dl)
	return buf
}
