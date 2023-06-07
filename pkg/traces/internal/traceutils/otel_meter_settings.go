package traceutils

import (
	"github.com/prometheus/client_golang/prometheus"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
)

func PromeheusExporter(reg prometheus.Registerer) (*otelprom.Exporter, error) {
	return otelprom.New(
		otelprom.WithRegisterer(reg),
		otelprom.WithoutUnits(),
		otelprom.WithoutScopeInfo(),
		otelprom.WithoutTargetInfo(),
		otelprom.WithNamespace("traces"))
}
