package msgs

type PushNotSubscribe struct {
	IgracId string `json:"igrac_id"`
	Uredjaj struct {
		GcmId   string `json:"gcm_id"`
		AppleId string `json:"apple_id"`
		Aktivan bool   `json:"aktivan"`
	} `json:"uredjaj"`
	Pretplate struct {
		PrivatnePoruke  bool `json:"privatne_poruke"`
		Novosti         bool `json:"novosti"`
		ListicVrednovan bool `json:"listic_vrednovan"`
	} `json:"pretplate"`
}
