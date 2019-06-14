package msgs

type UvjetiKoristenjaReq struct {
	IgracId  string `json:"igrac_id"`
	Verzija  int    `json:"verzija"`
	RemoteIP string `json:"remote_ip"`
}

type PostavkePrivatnostiReq struct {
	IgracId          string `json:"igrac_id"`
	NewsletterStatus int    `json:"newsletter_status"`
	SMSStatus        int    `json:"sms_status"`
	RemoteIP         string `json:"remote_ip"`
}
