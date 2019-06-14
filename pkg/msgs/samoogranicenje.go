package msgs

import "encoding/json"

type SamoogranicenjeSetReq struct {
	IgracId  string  `json:"igrac_id"`
	Iznos    float64 `json:"iznos"`
	BrojDana int     `json:"brojDana"`
	Tip      int     `json:"tip"`
}

func (req *SamoogranicenjeSetReq) ToJson() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

type SamoogranicenjeDelReq struct {
	Id      string `json:"id"`
	IgracId string `json:"igrac_id"`
}

func (req *SamoogranicenjeDelReq) ToJson() []byte {
	buf, _ := json.Marshal(req)
	return buf
}
