package msgs

import (
	"encoding/json"
	"lib/listic"
)

// Konstante potrebne za isplate listica i gotovinskih ostava
const (
	ListiciIsplatiStatusOk                 = 0
	ListiciIsplatiStatusIgracNijePronadjen = 1
	ListiciIsplatiStatusNijePronadjen      = 2
	ListiciIsplatiStatusNijeDobitan        = 3
	ListiciIsplatiStatusIsplacen           = 4
	ListiciIsplatiStatusNijeDozvoljeno     = 5
	ListiciIsplatiStatusNijeVrednovan      = 6
	ListiciIsplatiStatusStorniran          = 7

	ListiciIsplatiTipListic           = 1
	ListiciIsplatiTipGotovinskaOstava = 2
)

// ListiciIsplatiReq request za isplatu listica ili gotovinske ostave igracu na racun
type ListiciIsplatiReq struct {
	IgracID       string `json:"igrac_id"`
	Broj          string `json:"broj"`
	KontrolniBroj string `json:"kontrolni_broj"`
	Tip           int    `json:"tip"`
	PrijavaID     string `json:"prijava_id"`
}

// ToJSON request isplate u JSON
func (req *ListiciIsplatiReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// ListiciIsplatiRsp response isplate listica ili gotovinkse ostave igracu na racun
type ListiciIsplatiRsp struct {
	Status      int              `json:"status"`
	Raspolozivo float64          `json:"raspolozivo"`
	Dobitak     float64          `json:"dobitak"`
	Listic      *listic.Dokument `json:"listic,omitempty"`
	Tip         int              `json:"tip"`
}
