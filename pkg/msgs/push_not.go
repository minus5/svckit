package msgs

const tipListic = 3

type PushNot struct {
	Id       int `json:"listic_id"`
	GcmId    string
	AppleId  string
	FcmId    string
	FcmTopic string
	Tip      int
	Tekst    string
	Listic   *PushNotListic
}

type PushNotListic struct {
	Id      string
	Tip     int
	Status  int
	Dobitak float64
	Broj    string
}

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

func (l *PushNotListic) Serialize() map[string]interface{} {
	m := make(map[string]interface{})
	m["id"] = l.Id
	m["status"] = l.Status
	m["dobitak"] = l.Dobitak
	m["broj"] = l.Broj
	return m
}

func NewPushNotText(id int, tip int, gcmId, appleId string, tekst string) *PushNot {
	return &PushNot{Id: id, Tip: tip, GcmId: gcmId, AppleId: appleId, Tekst: tekst}
}

func NewPushNotListic(id int, tip int, lTip int, gcmId, appleId string, listicId string, status int, dobitak float64, broj string) *PushNot {
	pn := &PushNot{Id: id, Tip: tip, GcmId: gcmId, AppleId: appleId}
	if tip == tipListic {
		pn.Listic = &PushNotListic{Id: listicId, Tip: lTip, Status: status, Dobitak: dobitak, Broj: broj}
	}
	return pn
}

func (m *PushNot) IsGcm() bool {
	return m.GcmId != ""
}

func (m *PushNot) IsApn() bool {
	return m.AppleId != ""
}
