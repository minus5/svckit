package msgs

import "encoding/json"

// Podaci za poziv dodaj_listic na web_backendu
// TODO: Uklonit kad svi klijenti prijedju na dodavanje listica preko DBi-a
// Vidi file: listici_dodaj.go
type DodajListic struct {
	Broj          string `json:"broj"`
	KontrolniBroj string `json:"kontrolni_broj"`
	IgracId       string `json:"igrac_guid"`
}

func (dl *DodajListic) ToJson() []byte {
	buf, _ := json.Marshal(dl)
	return buf
}
