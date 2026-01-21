package msgs

import (
	"encoding/json"
	"log"
)

//Vrste uplate
const (
	vrstaUplateInternet = "internet"
	vrstaUplateTest     = "test"
	tipSportski         = "sportski"
)

//Listici citanje listica igraca
type Listici struct {
	Offset      int64  `json:"offset"`
	Limit       int64  `json:"limit"`
	VrstaUplate string `json:"vrsta_uplate"`
	Tip         string `json:"tip"`
}

type ListiciSeek struct {
	VrstaUplate     string `json:"vrsta_uplate"`
	Tip             string `json:"tip"`
	Status          int64  `json:"status"`
	StatusUpdatedAt string `json:"status_updated_at"`
	Time            string `json:"time"`
	Limit           int64  `json:"limit"`
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
	if l.Tip == "" {
		l.Tip = tipSportski
	}
	return l, nil
}

func ParseListiciSeek(body string, isTestIgrac bool) (*ListiciSeek, error) {
	l := &ListiciSeek{}
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
	if l.Tip == "" {
		l.Tip = tipSportski
	}
	return l, nil
}
