package def

type Metric interface {
	LogForCnt(int64)
}

type MetricWriter interface {
	AsyncWrite(metric Metric)
	Write(metric Metric) error
}
