package aws_firehose

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	errors          *prometheus.CounterVec
	entriesReceived *prometheus.CounterVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := metrics{}
	m.errors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loki_source_awsfirehose_parsing_errors",
		Help: "Number of parsing errors while receiving AWS Firehose messages",
	}, []string{"reason"})

	m.entriesReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loki_source_awsfirehose_entries_received",
		Help: "Number of entries received from AWS Firehose",
	}, []string{})

	if reg != nil {
		reg.MustRegister(
			m.errors,
			m.entriesReceived,
		)
	}

	return &m
}
