package file

import "github.com/prometheus/client_golang/prometheus"

// metrics hold the set of file-based metrics.
type metrics struct {
	// Registerer used. May be nil.
	reg prometheus.Registerer

	// File-specific metrics
	readBytes        *prometheus.GaugeVec
	totalBytes       *prometheus.GaugeVec
	readLines        *prometheus.CounterVec
	encodingFailures *prometheus.CounterVec
	filesActive      prometheus.Gauge
}

// newMetrics creates a new set of file metrics. If reg is non-nil, the metrics
// will be registered.
func newMetrics(reg prometheus.Registerer) *metrics {
	var m metrics
	m.reg = reg

	m.readBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "loki_source_file",
		Name:      "read_bytes_total",
		Help:      "Number of bytes read.",
	}, []string{"path"})
	m.totalBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "loki_source_file",
		Name:      "file_bytes_total",
		Help:      "Number of bytes total.",
	}, []string{"path"})
	m.readLines = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "loki_source_file",
		Name:      "read_lines_total",
		Help:      "Number of lines read.",
	}, []string{"path"})
	m.encodingFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "loki_source_file",
		Name:      "encoding_failures_total",
		Help:      "Number of encoding failures.",
	}, []string{"path"})
	m.filesActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "loki_source_file",
		Name:      "files_active_total",
		Help:      "Number of active files.",
	})

	if reg != nil {
		reg.MustRegister(
			m.readBytes,
			m.totalBytes,
			m.readLines,
			m.encodingFailures,
			m.filesActive,
		)
	}

	return &m
}
