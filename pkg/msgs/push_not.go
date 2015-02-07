package msgs

const tipListic = 3

type PushNot struct {
	Id      int
	GcmId   string
	AppleId string
	Tip     int
	Tekst   string
	Listic  *PushNotListic
}

type PushNotListic struct {
	Id      string
	Status  int
	Dobitak float64
	Broj    string
}

func (m *PushNot) Serialize() map[string]interface{} {
	d := make(map[string]interface{})
	d["tip"] = m.Tip
	if m.Listic != nil {
		l := make(map[string]interface{})
		l["id"] = m.Listic.Id
		l["status"] = m.Listic.Status
		l["dobitak"] = m.Listic.Dobitak
		l["broj"] = m.Listic.Broj
		d["listic"] = l
	}
	if m.Tekst != "" {
		d["tekst"] = m.Tekst
	}
	return d
}

func NewPushNotText(id int, tip int, gcmId, appleId string, tekst string) *PushNot {
	return &PushNot{Id: id, Tip: tip, GcmId: gcmId, AppleId: appleId, Tekst: tekst}
}

func NewPushNotListic(id int, tip int, gcmId, appleId string, listicId string, status int, dobitak float64, broj string) *PushNot {
	pn := &PushNot{Id: id, Tip: tip, GcmId: gcmId, AppleId: appleId}
	if tip == tipListic {
		pn.Listic = &PushNotListic{Id: listicId, Status: status, Dobitak: dobitak, Broj: broj}
	}
	return pn
}

func (m *PushNot) IsGcm() bool {
	return m.GcmId != ""
}

func (m *PushNot) IsApn() bool {
	return m.AppleId != ""
}
