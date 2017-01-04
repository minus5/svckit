package msgs

type DokumentiDopuna struct {
	IgracId    string `json:"igrac_id"`
	Tip        string `json:"tip"`
	Count      int64  `json:"count,omitempty"`
	MongoCount int64  `json:"mongo_count,omitempty"`
	Offset     int64  `json:"offset,omitempty"`
	Limit      int64  `json:"limit,omitempty"`
}

func (dd DokumentiDopuna) TipIgraci() bool {
	return dd.Tip == "igraci"
}
