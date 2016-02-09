package msgs

import "encoding/json"

type StatistikaIdRequest struct {
	Id      int `json:"id"`
	IzvorId int `json:"izvorId"`
	BaseId  int `json:"baseId"`
}

type StatistikaIdResponse struct {
	StatistikaId     int `json:"statistikaId"`
	StatistikaLigaId int `json:"statistikaLigaId"`
}

func (r *StatistikaIdRequest) ToJson() []byte {
	buf, _ := json.Marshal(r)
	return buf
}
