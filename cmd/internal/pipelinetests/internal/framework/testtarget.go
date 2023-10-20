package framework

import (
	"net/http/httptest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type TestTarget struct {
	registry *prometheus.Registry
	server   *httptest.Server
}

func newTestTarget() *TestTarget {
	target := &TestTarget{
		registry: prometheus.NewRegistry(),
	}

	server := httptest.NewServer(promhttp.InstrumentMetricHandler(
		target.registry, promhttp.HandlerFor(target.registry, promhttp.HandlerOpts{
			Registry: target.registry,
		}),
	))
	target.server = server

	return target
}

func (t *TestTarget) Register(collector prometheus.Collector) {
	t.registry.MustRegister(collector)
}
