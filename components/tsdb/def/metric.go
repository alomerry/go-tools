package def

type Metric interface {
	LogForCnt(int64)
	LogForLongVal(string, int64)
	LogForDoubleVal(string, float64)
}

type MetricWriter interface {
	AsyncWrite(metric Metric)
	Write(metric Metric) error
}
