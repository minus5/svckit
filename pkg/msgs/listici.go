package msgs

import (
	"encoding/json"
	"log"
)

type ListiciBrisiMessage struct {
	IgracId string   `json:"igrac_id"`
	Listici []string `json:"listici"`
}

type DokumentiDopuna struct {
	IgracId    string `json:"igrac_id"`
	Tip        string `json:"tip"`
	Count      int64  `json:"count"`
	MongoCount int64  `json:"mongo_count"`
	Offset     int64  `json:"offset"`
	Limit      int64  `json:"limit"`
}

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

//Vrste uplate
const (
	vrstaUplateInternet = "internet"
	vrstaUplateTest     = "test"
)

//Listici citanje listica igraca
type Listici struct {
	Offset      int64  `json:"offset"`
	Limit       int64  `json:"limit"`
	VrstaUplate string `json:"vrsta_uplate"`
}

//ParseListici parsiraj json
func ParseListici(body string, isTestIgrac bool) (*Listici, error) {
	l := &Listici{}
	if err := json.Unmarshal([]byte(body), l); err != nil {
		log.Printf("[ERROR] %s while parsing %s", err, body)
		return nil, err
	}
	if l.VrstaUplate == "" {
		l.VrstaUplate = vrstaUplateInternet
	}
	if l.VrstaUplate == vrstaUplateInternet && isTestIgrac {
		l.VrstaUplate = vrstaUplateTest
	}
	if l.Limit > 100 {
		l.Limit = 100
	}
	return l, nil
}
