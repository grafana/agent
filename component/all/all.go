// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/discovery/docker"                     // Import discovery.docker
	_ "github.com/grafana/agent/component/discovery/file"                       // Import discovery.file
	_ "github.com/grafana/agent/component/discovery/kubernetes"                 // Import discovery.kubernetes
	_ "github.com/grafana/agent/component/discovery/relabel"                    // Import discovery.relabel
	_ "github.com/grafana/agent/component/loki/source/file"                     // Import loki.source.file
	_ "github.com/grafana/agent/component/otelcol/auth/basic"                   // Import otelcol.auth.basic
	_ "github.com/grafana/agent/component/otelcol/auth/bearer"                  // Import otelcol.auth.bearer
	_ "github.com/grafana/agent/component/otelcol/auth/headers"                 // Import otelcol.auth.headers
	_ "github.com/grafana/agent/component/otelcol/exporter/otlp"                // Import otelcol.exporter.otlp
	_ "github.com/grafana/agent/component/otelcol/exporter/otlphttp"            // Import otelcol.exporter.otlphttp
	_ "github.com/grafana/agent/component/otelcol/exporter/prometheus"          // Import otelcol.exporter.prometheus
	_ "github.com/grafana/agent/component/otelcol/processor/batch"              // Import otelcol.processor.batch
	_ "github.com/grafana/agent/component/otelcol/processor/memorylimiter"      // Import otelcol.processor.memory_limiter
	_ "github.com/grafana/agent/component/otelcol/receiver/jaeger"              // Import otelcol.receiver.jaeger
	_ "github.com/grafana/agent/component/otelcol/receiver/otlp"                // Import otelcol.receiver.otlp
	_ "github.com/grafana/agent/component/otelcol/receiver/prometheus"          // Import otelcol.receiver.prometheus
	_ "github.com/grafana/agent/component/prometheus/integration/node_exporter" // Import prometheus.integration.node_exporter
	_ "github.com/grafana/agent/component/prometheus/relabel"                   // Import prometheus.relabel
	_ "github.com/grafana/agent/component/prometheus/remotewrite"               // Import prometheus.remote_write
	_ "github.com/grafana/agent/component/prometheus/scrape"                    // Import prometheus.scrape
	_ "github.com/grafana/agent/component/remote/http"                          // Import remote.http
	_ "github.com/grafana/agent/component/remote/s3"                            // Import remote.s3
)
