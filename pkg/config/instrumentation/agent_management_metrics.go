package instrumentation

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type agentManagementMetrics struct {
	fallbacks *prometheus.CounterVec
}

var amMetrics *agentManagementMetrics
var amMetricsInitializer sync.Once

func initializeagentManagementMetrics() {
	amMetrics = newAgentManagementMetrics()
}

func newAgentManagementMetrics() *agentManagementMetrics {
	var agentManagementMetrics agentManagementMetrics

	agentManagementMetrics.fallbacks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_management_fallbacks_total",
			Help: "Number of fallbacks by fallback destination.",
		},
		[]string{"destination"},
	)

	return &agentManagementMetrics
}

func InstrumentAgentManagementFallback(destination string) {
	amMetricsInitializer.Do(initializeagentManagementMetrics)
	amMetrics.fallbacks.WithLabelValues(destination).Inc()
}
