package crow

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	totalScrapes prometheus.Counter
	totalSamples prometheus.Counter
	totalResults *prometheus.CounterVec
	pendingSets  prometheus.Gauge

	cachedCollectors []prometheus.Collector
}

func newMetrics() *metrics {
	var m metrics

	m.totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "crow_test_scrapes_total",
		Help: "Total number of generated test sample sets",
	})

	m.totalSamples = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "crow_test_samples_total",
		Help: "Total number of generated test samples",
	})

	m.totalResults = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "crow_test_sample_results_total",
		Help: "Total validation results of test samples",
	}, []string{"result"}) // result is either "success", "missing", "mismatch", or "unknown"

	m.pendingSets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "crow_test_pending_validations",
		Help: "Total number of pending validations to perform",
	})

	return &m
}

func (m *metrics) collectors() []prometheus.Collector {
	if m.cachedCollectors == nil {
		m.cachedCollectors = []prometheus.Collector{
			m.totalScrapes,
			m.totalSamples,
			m.totalResults,
			m.pendingSets,
		}
	}
	return m.cachedCollectors
}

func (m *metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range m.collectors() {
		c.Describe(ch)
	}
}

func (m *metrics) Collect(ch chan<- prometheus.Metric) {
	for _, c := range m.collectors() {
		c.Collect(ch)
	}
}
