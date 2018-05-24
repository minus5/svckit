package msgs

type UvjetiKoristenjaReq struct {
	IgracId string `json:"igrac_id"`
	Verzija int    `json:"verzija"`
}

type NewsletterPostavkeReq struct {
	IgracId string `json:"igrac_id"`
	Status  int    `json:"status"`
}
