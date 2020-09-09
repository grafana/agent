package tempo

import (
	"contrib.go.opencensus.io/exporter/prometheus"
	prom_client "github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/obsreport"
)

// applicationTelemetry is application's own telemetry.
var applicationTelemetry appTelemetryExporter = &appTelemetry{}

type appTelemetryExporter interface {
	init(logger *zap.Logger) error
	shutdown() error
}

type appTelemetry struct {
	views []*view.View
}

func (tel *appTelemetry) init(logger *zap.Logger) error {
	var views []*view.View
	views = append(views, obsreport.Configure(false, true)...)
	tel.views = views
	if err := view.Register(views...); err != nil {
		return err
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace:  "tempo",
		Registerer: prom_client.DefaultRegisterer,
	})
	if err != nil {
		return err
	}

	view.RegisterExporter(pe)

	return nil
}

func (tel *appTelemetry) shutdown() error {
	view.Unregister(tel.views...)

	return nil
}
