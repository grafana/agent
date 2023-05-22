package internal

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	errors          *prometheus.CounterVec
	recordsReceived *prometheus.CounterVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := metrics{}
	m.errors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loki_source_awsfirehose_errors",
		Help: "Number of errors while receiving AWS Firehose messages. This includes things like reading the HTTP body, json decoding, etc.",
	}, []string{"reason"})

	m.recordsReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loki_source_awsfirehose_records_received",
		Help: "Number of records received from AWS Firehose",
	}, []string{"type"})

	if reg != nil {
		reg.MustRegister(
			m.errors,
			m.recordsReceived,
		)
	}

	return &m
}
