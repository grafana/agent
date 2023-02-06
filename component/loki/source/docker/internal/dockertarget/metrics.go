package dockertarget

// This code is copied from Promtail. The dockertarget package is used to
// configure and run the targets that can read logs from Docker containers and
// forward them to other loki components.

import "github.com/prometheus/client_golang/prometheus"

// Metrics holds a set of Docker target metrics.
type Metrics struct {
	reg prometheus.Registerer

	dockerEntries prometheus.Counter
	dockerErrors  prometheus.Counter
}

// NewMetrics creates a new set of Docker target metrics. If reg is non-nil, the
// metrics will be registered.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	var m Metrics
	m.reg = reg

	m.dockerEntries = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loki_source_docker_target_entries_total",
		Help: "Total number of successful entries sent to the Docker target",
	})
	m.dockerErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loki_source_docker_target_parsing_errors_total",
		Help: "Total number of parsing errors while receiving Docker messages",
	})

	if reg != nil {
		reg.MustRegister(
			m.dockerEntries,
			m.dockerErrors,
		)
	}

	return &m
}
