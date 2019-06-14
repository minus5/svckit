package msgs

import (
	"encoding/json"
	"log"
)

//Transakcije citanje transakcija igraca
type Transakcije struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

//ParseTransakcije parsiraj json
func ParseTransakcije(body string) (*Transakcije, error) {
	t := &Transakcije{}
	if err := json.Unmarshal([]byte(body), t); err != nil {
		log.Printf("[ERROR] %s while parsing %s", err, body)
		return nil, err
	}
	return t, nil
}
