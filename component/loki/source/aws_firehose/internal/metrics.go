package internal

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	errors          *prometheus.CounterVec
	recordsReceived *prometheus.CounterVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := metrics{}
	m.errors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loki_source_awsfirehose_parsing_errors",
		Help: "Number of parsing errors while receiving AWS Firehose messages",
	}, []string{"reason"})

	m.recordsReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loki_source_awsfirehose_records_received",
		Help: "Number of records received from AWS Firehose",
	}, []string{})

	if reg != nil {
		reg.MustRegister(
			m.errors,
			m.recordsReceived,
		)
	}

	return &m
}
