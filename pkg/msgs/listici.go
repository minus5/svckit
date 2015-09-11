package msgs

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
