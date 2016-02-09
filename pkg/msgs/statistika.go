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

func (r StatistikaIdRequest) ToJson() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func (r StatistikaIdResponse) ToJson() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func StatistikaRspFromJson(b []byte) (*StatistikaIdResponse, error) {
	rsp := &StatistikaIdResponse{}
	err := json.Unmarshal(b, rsp)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func StatistikaReqFromJson(b []byte) (*StatistikaIdRequest, error) {
	req := &StatistikaIdRequest{}
	err := json.Unmarshal(b, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
