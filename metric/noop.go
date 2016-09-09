package metric

// Do nothing implemnetation of Metric interface.
type Noop struct{}

func (*Noop) Counter(name string, values ...int) {
	return
}

func (*Noop) Gauge(name string, value int) {
	return
}

func (*Noop) Timing(name string, f func()) {
	//sw := NewStopwatch()
	f()
	//log.Printf("timer %s %d ms", name, sw.GetNs())
}

func (*Noop) Time(name string, duration int) {
	return
}

func (*Noop) WithPrefix(prefix string) Metric {
	return &Noop{}
}

func (*Noop) AppendSuffix(suffix string) Metric {
	return &Noop{}
}

func NewNoop() *Noop {
	return &Noop{}
}
