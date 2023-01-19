package instrumentation

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type remoteConfigMetrics struct {
	fetchStatusCodes *prometheus.CounterVec
	fetchErrors      prometheus.Counter
}

var metrics *remoteConfigMetrics
var metricsInitializer sync.Once

func initializeRemoteConfigMetrics() {
	metrics = newRemoteConfigMetrics()
}

func newRemoteConfigMetrics() *remoteConfigMetrics {
	var remoteConfigMetrics remoteConfigMetrics

	remoteConfigMetrics.fetchStatusCodes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "remote_config_fetches_total",
			Help: "Number of fetch requests for the remote config by HTTP status code",
		},
		[]string{"code"},
	)
	remoteConfigMetrics.fetchErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "remote_config_fetch_errors_total",
			Help: "Number of errors attempting to fetch remote config",
		},
	)

	return &remoteConfigMetrics
}

func InstrumentRemoteConfigFetch(code int) {
	metricsInitializer.Do(initializeRemoteConfigMetrics)
	metrics.fetchStatusCodes.WithLabelValues(fmt.Sprintf("%d", code)).Inc()
}

func InstrumentRemoteConfigFetchError() {
	metricsInitializer.Do(initializeRemoteConfigMetrics)
	metrics.fetchErrors.Inc()
}
