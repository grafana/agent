package instrumentation

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type remoteConfigMetrics struct {
	fetchStatusCodes *prometheus.CounterVec
	fetchErrors      prometheus.Counter
}

var RemoteConfigMetrics = newRemoteConfigMetrics()

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

func (r *remoteConfigMetrics) InstrumentRemoteConfigFetch(code int) {
	r.fetchStatusCodes.WithLabelValues(fmt.Sprintf("%d", code)).Inc()
}

func (r *remoteConfigMetrics) InstrumentRemoteConfigFetchError() {
	r.fetchErrors.Inc()
}
