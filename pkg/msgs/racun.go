package msgs

import (
	"encoding/json"
)

// IsplataPoslovnicaReq request za isplatu u poslovnici
type IsplataPoslovnicaReq struct {
	ID            string `json:"guid"`
	RememberToken string `json:"remember_token"`
	Iznos         string `json:"iznos"`
	Poslovnica    string `json:"poslovnica"`
}

// ToJSON ...
func (req *IsplataPoslovnicaReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// IsplataRacunReq request za isplatu na racun
type IsplataRacunReq struct {
	RememberToken string `json:"remember_token"`
	Iznos         string `json:"iznos"`
	Racun         string `json:"racun"`
	Iban          string `json:"iban"`
}

// ToJSON ...
func (req *IsplataRacunReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// IsplataSkrillReq request za isplatu na skrill racun
type IsplataSkrillReq struct {
	RememberToken string `json:"remember_token"`
	Email         string `json:"email"`
	Type          string `json:"gateway_type"`
	Iznos         string `json:"iznos"`
}

// ToJSON ...
func (req *IsplataSkrillReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// ObrisiRacunReq request za brisanje tekucih racuna
type ObrisiRacunReq struct {
	RememberToken string `json:"remember_token"`
	RacunID       string `json:"id"`
}

// ToJSON ...
func (req *ObrisiRacunReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// UplatiNaRacunReq request za uplatu na racun
type UplatiNaRacunReq struct {
	RememberToken string `json:"remember_token"`
	GatewayType   string `json:"gateway_type"`
	PaymentType   string `json:"payment_type"`
	Email         string `json:"email"`
	Iznos         string `json:"iznos"`
}

// ToJSON ...
func (req *UplatiNaRacunReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// PromijeniLozinkuReq request za izmjenu lozinke
type PromijeniLozinkuReq struct {
	RememberToken  string `json:"remember_token"`
	StaraLozinka   string `json:"old_password"`
	Lozinka        string `json:"password"`
	LozinkaPotvrda string `json:"password_confirmation"`
}

// ToJSON ...
func (req *PromijeniLozinkuReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// PromijeniEmailReq request za izmjenu email adrese
type PromijeniEmailReq struct {
	RememberToken string `json:"remember_token"`
	Email         string `json:"email"`
	Lozinka       string `json:"password"`
}

// ToJSON ...
func (req *PromijeniEmailReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}

// PromijeniEmailReq request za izmjenu email adrese
type PromijeniPodatkeReq struct {
	RememberToken string `json:"remember_token"`
	Adresa        string `json:"adresa,omitempty"`
	Grad          string `json:"grad,omitempty"`
	PostanskiBroj string `json:"postanski_broj,omitempty"`
	Telefon       string `json:"telefon,omitempty"`
}

// ToJSON ...
func (req *PromijeniPodatkeReq) ToJSON() []byte {
	buf, _ := json.Marshal(req)
	return buf
}
