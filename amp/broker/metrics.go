package broker

type metricer struct {
	Time func(string, int)
}

var metric = metricer{
	Time: func(string, int) {},
}

func UseMetric(time func(string, int)) {
	metric.Time = time
}
