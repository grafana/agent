package instrumentation

import (
	"crypto/sha256"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type configMetrics struct {
	configHash          *prometheus.GaugeVec
	configReloadSuccess prometheus.Gauge
	configReloadSeconds *prometheus.GaugeVec
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
	m.configReloadSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "agent_config_last_reload_successful",
		Help: "Config loaded successfully.",
	})
	m.configReloadSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "agent_config_last_reload_timestamp_seconds",
		Help: "Timestamp of the last configuration reload by result.",
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

// Expose metrics for reload success / failures.
func (c *configMetrics) InstrumentLoad(isError bool) {
	if isError {
		c.configReloadSuccess.Set(0)
		c.configReloadSeconds.WithLabelValues("failure").SetToCurrentTime()
	} else {
		c.configReloadSuccess.Set(1)
		c.configReloadSeconds.WithLabelValues("success").SetToCurrentTime()
	}
}
