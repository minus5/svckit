package msgs

type ListiciBrisiMessage struct {
	IgracId string   `json:"igrac_id"`
	Listici []string `json:"listici"`
}

type ListiciDopuna struct {
	Tip        string `json:"tip"`
	Count      int    `json:"count"`
	MongoCount int    `json:"mongo_count"`
	Offset     int    `json:"offset"`
	Limit      int    `json:"limit"`
}
