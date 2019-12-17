package broker

type metricer struct {
	Time func(string, int)
	Count
}

var metric = metricer{
	Time: func(string, int) {},
}

func UseMetric(time func(string, int)) {
	metric.Time = time
}
