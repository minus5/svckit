package msgs

import (
	"encoding/json"
)

type SportradarIdRequest struct {
	Id      int    `json:"id"`
	Izvor   int    `json:"izvor"`
	IzvorId int    `json:"izvorId"`
	BaseId  int    `json:"baseId"`
	Status  string `json:"status"`
}

type LigaInfoRequest struct {
	BaseId int `json:"baseId"`
}

type LigaInfoResponse struct {
	LigaId         int     `json:"ligaId"`
	LigaNaziv      string  `json:"ligaNaziv"`
	LigaGrupaId    *int    `json:"ligaGrupaId"`
	LigaGrupaNaziv *string `json:"ligaGrupaNaziv"`
}

type SpecijalInfoResponse struct {
	Id int
}

type SportradarIdResponse struct {
	SportradarId     int `json:"statistikaId"`
	SportradarLigaId int `json:"statistikaLigaId"`
}

func (r SportradarIdRequest) ToJson() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func (r SportradarIdResponse) ToJson() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func (r LigaInfoRequest) ToJson() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func (r LigaInfoResponse) ToJson() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func SportradarRspFromJson(b []byte) (*SportradarIdResponse, error) {
	rsp := &SportradarIdResponse{}
	err := json.Unmarshal(b, rsp)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func LigaInfoRspFromJson(b []byte) (*LigaInfoResponse, error) {
	rsp := &LigaInfoResponse{}
	err := json.Unmarshal(b, rsp)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func SportradarReqFromJson(b []byte) (*SportradarIdRequest, error) {
	req := &SportradarIdRequest{}
	err := json.Unmarshal(b, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func LigaInfoReqFromJson(b []byte) (*LigaInfoRequest, error) {
	req := &LigaInfoRequest{}
	err := json.Unmarshal(b, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

const (
	SportradarStatusNepoznat  = "nepoznat"
	SportradarStatusNovi      = "novi dogadjaj"
	SportradarStatusAktiviran = "dogadjaj aktiviran"
	SportradarStatusBaseId    = "postavljen base ID"
)
