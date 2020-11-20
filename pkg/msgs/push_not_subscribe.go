package msgs

// Tipovi uredjaja koje podrzavamo
const (
	PushNotDeviceTypeAndroid = iota
	PushNotDeviceTypeiOS
	PushNotDeviceTypeWeb
)

// Poruka za prijavu na push notifikacije
type PushNotSubscribe struct {
	// remember_token igraca koji se prijavljuje za notifikacije
	IgracId string `json:"igrac_id"`
	// Uredjaj s kojim se prijavljuje za notifikacije
	Uredjaj struct {
		// Firebase Cloud Messaging device id
		FcmId string `json:"fcm_id"`
		// Da li je uredjaj aktivan, ako je false koristi se za deaktivaciju uredjaja
		Aktivan bool `json:"aktivan"`
		// Tip uredjaja za primanje push notifikacija, 0 - Android, 1 - iOS, 2 - web
		DeviceType int `json:"device_type"`
	} `json:"uredjaj"`
	// Notifikacije na koje se igrac pretplacuje
	Pretplate struct {
		PrivatnePoruke  bool `json:"privatne_poruke"`
		Novosti         bool `json:"novosti"`
		ListicVrednovan bool `json:"listic_vrednovan"`
		ParVrednovan    bool `json:"par_vrednovan"`
	} `json:"pretplate"`
}
