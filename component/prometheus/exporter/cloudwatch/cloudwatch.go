package cloudwatch

import (
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/cloudwatch_exporter"
	"github.com/grafana/agent/service/http"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.cloudwatch",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: []string{http.ServiceName},
		Build:         exporter.New(createExporter, "cloudwatch"),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	exporterConfig, err := ConvertToYACE(a)
	if err != nil {
		return nil, fmt.Errorf("invalid cloudwatch exporter configuration: %w", err)
	}
	// yaceSess expects a default value of True
	fipsEnabled := !a.FIPSDisabled

	return cloudwatch_exporter.NewCloudwatchExporter(opts.ID, opts.Logger, exporterConfig, fipsEnabled, a.Debug), nil
}
