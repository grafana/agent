package runner

import (
	"bytes"
	"testing"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/pkg/autodiscovery"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	input := []*autodiscovery.Result{
		{
			RiverConfig: `prometheus.exporter.mysql "default" {
  data_source_name = "user@host"
}`,
			MetricsExport: "prometheus.exporter.mysql.default.targets",
			LogfileTargets: []discovery.Target{
				{"__path__": "/tmp/logs/mysql.log", "lbl": "mysql"},
			},
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

	expected := `prometheus.scrape "default" {
	targets = concat(
		prometheus.exporter.mysql.default.targets,
		prometheus.exporter.consul.default.targets,
		[
			{__address__ = "localhost:9090", lbl = "foo"},
			{__address__ = "localhost:9091", lbl = "bar"},
		],
	)
	forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
	endpoint {
		url = env("GRAFANACLOUD_METRICS_URL")

		basic_auth {
			username = env("GRAFANACLOUD_METRICS_USER")
			password = env("GRAFANACLOUD_APIKEY")
		}
	}
}

loki.source.file "default" {
	targets = [
		{__path__ = "/tmp/logs/mysql.log", lbl = "mysql"},
		{__path__ = "/tmp/logs/1.log", lbl = "foo"},
		{__path__ = "/tmp/logs/2.log", lbl = "bar"},
	]

	forward_to = [loki.write.default.receiver]
}

loki.write "default" {
	endpoint {
		url = env("GRAFANACLOUD_LOGS_URL")

		basic_auth {
			username = env("GRAFANACLOUD_LOGS_USER")
			password = env("GRAFANACLOUD_APIKEY")
		}
	}
}

prometheus.exporter.mysql "default" {
	data_source_name = "user@host"
}

prometheus.exporter.consul "default" {
	server = "https://consul.example.com:8500"
}`

	buf := new(bytes.Buffer)
	RenderConfig(buf, BuildTemplateInput(input))

	require.Equal(t, expected, buf.String())
}
