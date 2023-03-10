package runner

import (
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
)

func Test(t *testing.T) {
	input := []*autodiscovery.Result{
		{
			RiverConfig: `prometheus.exporter.mysql "default" {
  data_source_name = "user@host"
}`,
			MetricsExport: "prometheus.exporter.mysql.default.targets",
		},
		{
			RiverConfig: `prometheus.exporter.consul "default" {
  server = "https://consul.example.com:8500"
}`,
			MetricsExport: "prometheus.exporter.consul.default.targets",
		},
		{
			MetricsTargets: []discovery.Target{
				{"__address__": "localhost:9090", "lbl": "foo"},
				{"__address__": "localhost:9091", "lbl": "bar"},
			},
		},
		{
			LogfileTargets: []discovery.Target{
				{"__path__": "/tmp/logs/1.log", "lbl": "foo"},
				{"__path__": "/tmp/logs/2.log", "lbl": "bar"},
			},
		},
	}

	RenderConfig(BuildTemplateInput(input))
}
