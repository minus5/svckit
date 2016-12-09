package msgs

import (
	"encoding/json"
	"log"
)

const (
	tip      = "all"
	fromDate = "begin"
	toDate   = "end"
)

//Transakcije citanje transakcija igraca
type Transakcije struct {
	Offset   int64  `json:"offset"`
	Limit    int64  `json:"limit"`
	Tip      string `json:"tip"`
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
}

//ParseTransakcije parsiraj json
func ParseTransakcije(body string) (*Transakcije, error) {
	t := &Transakcije{}
	if err := json.Unmarshal([]byte(body), t); err != nil {
		log.Printf("[ERROR] %s while parsing %s", err, body)
		return nil, err
	}
	if t.Tip == "" {
		t.Tip = tip
	}
	if t.FromDate == "" {
		t.FromDate = fromDate
	}
	if t.ToDate == "" {
		t.ToDate = toDate
	}

	return t, nil
}
