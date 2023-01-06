package instrumentation

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type remoteConfigMetrics struct {
	fetchStatusCodes *prometheus.CounterVec
}

var RemoteConfigMetrics = newRemoteConfigMetrics()

func newRemoteConfigMetrics() *remoteConfigMetrics {
	var remoteConfigMetrics remoteConfigMetrics

	remoteConfigMetrics.fetchStatusCodes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "remote_config_fetches",
			Help: "Counter of fetch requests for the remote config by HTTP status code",
		},
		[]string{"code"},
	)
	return &remoteConfigMetrics
}

func (r *remoteConfigMetrics) RemoteConfigFetch(code int) {
	r.fetchStatusCodes.WithLabelValues(fmt.Sprintf("%d", code)).Inc()
}
