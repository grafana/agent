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

func (args *traceConfig) convert() map[string]interface{} {
	if args == nil {
		return nil
	}

	result := make(map[string]interface{})
	if len(args.Span) > 0 {
		result["span"] = append([]string{}, args.Span...)
	}
	if len(args.SpanEvent) > 0 {
		result["spanevent"] = append([]string{}, args.SpanEvent...)
	}

	return result
}

func (args *metricConfig) convert() map[string]interface{} {
	if args == nil {
		return nil
	}

	result := make(map[string]interface{})
	if len(args.Metric) > 0 {
		result["metric"] = append([]string{}, args.Metric...)
	}
	if len(args.Datapoint) > 0 {
		result["datapoint"] = append([]string{}, args.Datapoint...)
	}

	return result
}

func (args *logConfig) convert() map[string]interface{} {
	if args == nil {
		return nil
	}

	return map[string]interface{}{
		"log_record": append([]string{}, args.LogRecord...),
	}
}
