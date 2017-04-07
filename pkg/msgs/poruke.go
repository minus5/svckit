package msgs

import (
	"encoding/json"
	"log"
)

//Poruke citanje poruke igraca
type Poruke struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

//ParsePoruke parsiraj json
func ParsePoruke(body string) (*Poruke, error) {
	t := &Poruke{}
	if err := json.Unmarshal([]byte(body), t); err != nil {
		log.Printf("[ERROR] %s while parsing %s", err, body)
		return nil, err
	}
	return t, nil
}
