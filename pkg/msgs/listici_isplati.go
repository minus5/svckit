package msgs

import "encoding/json"

const (
	ListiciIsplatiStatusOk = iota
	ListiciIsplatiStatusIgracNijePronadjen
	ListiciIsplatiStatusNijePronadjen
	ListiciIsplatiStatusNijeDobitan
	ListiciIsplatiStatusIsplacen
)

type ListiciIsplatiReq struct {
	IgracId       string `json:"igrac_id"`
	Broj          string `json:"broj"`
	KontrolniBroj string `json:"kontrolni_broj"`
}

type ListiciIsplatiRsp struct {
	Status      int
	Raspolozivo float64
	Dobitak     float64
	Listic      map[string]interface{}
}

func (req *ListiciIsplatiReq) ToJson() []byte {
	buf, _ := json.Marshal(req)
	return buf
}
