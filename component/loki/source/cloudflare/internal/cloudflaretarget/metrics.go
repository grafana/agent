package cloudflaretarget

// This code is copied from Promtail. The cloudflaretarget package is used to
// configure and run a target that can read from the Cloudflare Logpull API and
// forward entries to other loki components.

import "github.com/prometheus/client_golang/prometheus"

// Metrics holds a set of cloudflare metrics.
type Metrics struct {
	reg prometheus.Registerer

	Entries prometheus.Counter
	LastEnd prometheus.Gauge
}

// NewMetrics creates a new set of cloudflare metrics. If reg is non-nil, the
// metrics will be registered.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	var m Metrics
	m.reg = reg

	m.Entries = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loki_source_cloudflare_target_entries_total",
		Help: "Total number of successful entries sent via the cloudflare target.",
	})
	m.LastEnd = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "loki_source_cloudflare_target_last_requested_end_timestamp",
		Help: "The last cloudflare request end timestamp fetched. This allows to calculate how far the target is behind.",
	})

	if reg != nil {
		reg.MustRegister(
			m.Entries,
			m.LastEnd,
		)
	}

	return &m
}
