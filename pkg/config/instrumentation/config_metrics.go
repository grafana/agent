package instrumentation

import (
	"crypto/sha256"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type configMetrics struct {
	configHash        *prometheus.GaugeVec
	configLoadSuccess prometheus.Gauge
	configLoadSeconds *prometheus.GaugeVec
}

// ConfigMetrics exposes metrics related to configuration loading
var ConfigMetrics = newConfigMetrics()

func newConfigMetrics() *configMetrics {
	var m configMetrics

	m.configHash = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_config_hash",
			Help: "Hash of the currently active config file.",
		},
		[]string{"sha256"},
	)
	m.configLoadSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "agent_config_last_load_successful",
		Help: "Config loaded successfully.",
	})
	m.configLoadSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "agent_config_last_load_timestamp_seconds",
		Help: "Timestamp of the last configuration load by result.",
	}, []string{"result"})

	return &m
}

// Create a sha256 hash of the config before expansion and expose it via
// the agent_config_hash metric.
func (c *configMetrics) InstrumentConfig(buf []byte) {
	hash := sha256.Sum256(buf)
	c.configHash.Reset()
	c.configHash.WithLabelValues(fmt.Sprintf("%x", hash)).Set(1)
}

// Expose metrics for load success / failures.
func (c *configMetrics) InstrumentLoad(success bool) {
	if success {
		c.configLoadSuccess.Set(1)
		c.configLoadSeconds.WithLabelValues("success").SetToCurrentTime()
	} else {
		c.configLoadSuccess.Set(0)
		c.configLoadSeconds.WithLabelValues("failure").SetToCurrentTime()
	}
}
