package client

import "github.com/prometheus/client_golang/prometheus"

type QueueClientMetrics struct {
	lastReadTimestamp *prometheus.GaugeVec
}

func NewQueueClientMetrics(reg prometheus.Registerer) *QueueClientMetrics {
	m := &QueueClientMetrics{
		lastReadTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "loki_write",
				Name:      "last_read_timestamp",
				Help:      "Latest timestamp read from the WAL",
			},
			[]string{"id"},
		),
	}

	if reg != nil {
		reg.MustRegister(m.lastReadTimestamp)
	}

	return m
}

func (m *QueueClientMetrics) CurryWithId(id string) *QueueClientMetrics {
	return &QueueClientMetrics{
		lastReadTimestamp: m.lastReadTimestamp.MustCurryWith(map[string]string{
			"id": id,
		}),
	}
}
