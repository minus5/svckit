package msgs

import "encoding/json"

const (
	ListiciIsplatiStatusOk                 = 0
	ListiciIsplatiStatusIgracNijePronadjen = 1
	ListiciIsplatiStatusNijePronadjen      = 2
	ListiciIsplatiStatusNijeDobitan        = 3
	ListiciIsplatiStatusIsplacen           = 4

	ListiciIsplatiTipListic           = 1
	ListiciIsplatiTipGotovinskaOstava = 2
)

type ListiciIsplatiReq struct {
	IgracId       string `json:"igrac_id"`
	Broj          string `json:"broj"`
	KontrolniBroj string `json:"kontrolni_broj"`
}

type ListiciIsplatiRsp struct {
	Status      int                    `json:"status"`
	Raspolozivo float64                `json:"raspolozivo"`
	Dobitak     float64                `json:"dobitak"`
	Listic      map[string]interface{} `json:"listic,omitempty"`
	Tip         int                    `json:"tip"`
}

func (req *ListiciIsplatiReq) ToJson() []byte {
	buf, _ := json.Marshal(req)
	return buf
}
