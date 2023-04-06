package instrumentation

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type agentManagementMetrics struct {
	configFallbacks *prometheus.CounterVec
}

var amMetrics *agentManagementMetrics
var amMetricsInitializer sync.Once

func initializeAgentManagementMetrics() {
	amMetrics = newAgentManagementMetrics()
}

func newAgentManagementMetrics() *agentManagementMetrics {
	var agentManagementMetrics agentManagementMetrics

	agentManagementMetrics.configFallbacks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_management_config_fallbacks_total",
			Help: "Number of config fallbacks by fallback destination.",
		},
		[]string{"destination"},
	)

	return &agentManagementMetrics
}

func InstrumentAgentManagementConfigFallback(destination string) {
	amMetricsInitializer.Do(initializeAgentManagementMetrics)
	amMetrics.configFallbacks.WithLabelValues(destination).Inc()
}
