package monitor

type OrderCreateMetric interface {
	Success(price int64)
	Failed(err error)
	PublishFailed()
}

type OrderMetrics interface {
	Create() OrderCreateMetric
}
