package filter

type traceConfig struct {
	Span      []string `river:"span,attr,optional"`
	SpanEvent []string `river:"spanevent,attr,optional"`
}

type metricConfig struct {
	Metric    []string `river:"metric,attr,optional"`
	Datapoint []string `river:"datapoint,attr,optional"`
}
type logConfig struct {
	LogRecord []string `river:"log_record,attr,optional"`
}
