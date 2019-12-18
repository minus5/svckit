package broker

type metricer struct {
	Time    func(string, int)
	Counter func(string, ...int)
}

var metric = metricer{
	Time:    func(string, int) {},
	Counter: func(s string, i ...int) {},
}

func UseMetric(time func(string, int), counter func(string, ...int)) {
	metric.Time = time
	metric.Counter = counter
}
