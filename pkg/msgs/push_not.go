package msgs

// Tipovi push not poruka
const (
	PushNotMsgTipPrivatna	= 1
	PushNotMsgTipBroadcast	= 2
	PushNotMsgTipListic		= 3
)

// Poruka za slanje na push notifikacije
type PushNot struct {
	Id			int `json:"push_not_id"`
	GcmId		string
	AppleId		string
	FcmId		string
	FcmTopic	string
	DeviceType	int
	Tip			int
	Tekst		string
	Listic		*PushNotListic
}

// Podaci listica za slanje na push notifikacije
type PushNotListic struct {
	Id      string
	Tip     int
	Status  int
	Dobitak float64
	Broj    string
}

// Serializira poruku i pretvara ju u poruku koja se salje na push notifikacije
func (m *PushNot) Serialize() map[string]interface{} {
	d := make(map[string]interface{})
	d["tip"] = m.Tip
	if m.Listic != nil {
		l := m.Listic.Serialize()
		d["listic"] = l
	}
	if m.Tekst != "" {
		d["tekst"] = m.Tekst
	}
	return d
}

// Serializira listic koji se salje kao poruka na push notifikacije
func (l *PushNotListic) Serialize() map[string]interface{} {
	m := make(map[string]interface{})
	m["id"] = l.Id
	m["status"] = l.Status
	m["dobitak"] = l.Dobitak
	m["broj"] = l.Broj
	return m
}

// Kreira novu tekstualnu push notification poruku
func NewPushNotText(id int, tip int, gcmId, appleId, fcmId string, deviceType int, tekst string) *PushNot {
	return &PushNot{Id: id, Tip: tip, GcmId: gcmId, AppleId: appleId, FcmId: fcmId, DeviceType: deviceType, Tekst: tekst}
}

// Kreira novu push notification poruku za status listica
func NewPushNotListic(id int, tip int, lTip int, gcmId, appleId, fcmId string, deviceType int, listicId string, status int, dobitak float64, broj string) *PushNot {
	pn := &PushNot{Id: id, Tip: tip, GcmId: gcmId, AppleId: appleId, FcmId: fcmId, DeviceType: deviceType}
	if tip == PushNotMsgTipListic {
		pn.Listic = &PushNotListic{Id: listicId, Tip: lTip, Status: status, Dobitak: dobitak, Broj: broj}
	}
	return pn
}

// Da li je poruka za FCM klijent ili FCM topic poruka
func (m *PushNot) IsFcm() bool {
	return m.FcmId != "" || m.FcmTopic != ""
}

// Da li je poruka za GCM klijent
func (m *PushNot) IsGcm() bool {
	return m.GcmId != ""
}

// Da li je poruka za iOS klijent
func (m *PushNot) IsApn() bool {
	return m.AppleId != ""
}
