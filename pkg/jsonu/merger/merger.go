package merger

// Metric interface koji opisuje sto treba imati metric
// inicijalno koristimo noopMetric koji ne radi nista
type Metric interface {
	Counter(name string, values ...int)
}

type noopMetric struct{}

func (*noopMetric) Counter(name string, values ...int) {
	return
}

var mtrc Metric

func init() {
	mtrc = &noopMetric{}
}

// SetMetric postavlja 'pravu' imeplementaciju metric-a
func SetMetric(m Metric) {
	mtrc = m
}

//limitira koliko ce se out of order diffova cekati za pojedini kanal prije trazenja dopune
var oooLimits map[string]int64 = make(map[string]int64)

func SetOOOLimit(kanal string, limit int64) {
	oooLimits[kanal]=limit
}