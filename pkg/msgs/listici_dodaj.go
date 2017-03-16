package msgs

import (
	"encoding/json"
	"lib/listic"
)

// Konstante za dodavanje listica igracu na racun
const (
	ListiciDodajStatusNijeMoguce            = -2 //pogodio je broj i kontrolni broj ali ne moze loto listic, ili konji ili web -- deprecated
	ListiciDodajStatusNijePronadjen         = -1
	ListiciDodajStatusOk                    = 0
	ListiciDodajStatusDodjeljenDrugomIgracu = 1
	ListiciDodajStatusIgracNijePronadjen    = 2
)

// ListiciDodajReq request podaci za dodavanje listica igracu
type ListiciDodajReq struct {
	Broj          string `json:"broj"`
	KontrolniBroj string `json:"kontrolni_broj"`
	IgracID       string `json:"igrac_id"`
}

// ListiciDodajRsp response podaci nakon dodavanja listica igracu
type ListiciDodajRsp struct {
	Status int              `json:"status"`
	Listic *listic.Dokument `json:"listic,omitempty"`
}

// ToJSON ...
func (req *ListiciDodajReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// IsplataStornoReq request za storniranje transakcije
type IsplataStornoReq struct {
	ID      string `json:"id"`
	IgracID string `json:"igrac_id"`
}

// ToJSON ...
func (req *IsplataStornoReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}
