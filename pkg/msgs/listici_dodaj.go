package msgs

import "encoding/json"

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
	IgracId       string `json:"igrac_id"`
}

// ListiciDodajRsp response podaci nakon dodavanja listica igracu
type ListiciDodajRsp struct {
	Status int                    `json:"status"`
	Listic map[string]interface{} `json:"listic,omitempty"`
}

// ToJson ...
func (req *ListiciDodajReq) ToJson() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// IsplataStornoReq request za storniranje transakcije
type IsplataStornoReq struct {
	Id      string `json:"id"`
	IgracId string `json:"igrac_id"`
}

// ToJson ...
func (req *IsplataStornoReq) ToJson() []byte {
	buf, _ := json.Marshal(req)
	return buf
}
