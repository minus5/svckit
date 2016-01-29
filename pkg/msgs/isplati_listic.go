package msgs

type IsplatiListicReq struct {
	IgracId string `json:"igrac_id"`
	Broj    string `json:"broj"`
	Kod     string `json:"kod"`
}

type IsplatiListicRsp struct {
	Status      int
	Raspolozivo float64
	Dobitak     float64
	Listic      map[string]interface{}
}

const (
	IsplatiListicStatusOk = iota
	IsplatiListicStatusIgracNijePronadjen
	IsplatiListicStatusNijePronadjen
	IsplatiListicStatusNijeDobitan
	IsplatiListicStatusIsplacen
)
