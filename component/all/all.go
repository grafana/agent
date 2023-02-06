// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/discovery/docker"                     // Import discovery.docker
	_ "github.com/grafana/agent/component/discovery/file"                       // Import discovery.file
	_ "github.com/grafana/agent/component/discovery/kubernetes"                 // Import discovery.kubernetes
	_ "github.com/grafana/agent/component/discovery/relabel"                    // Import discovery.relabel
	_ "github.com/grafana/agent/component/local/file"                           // Import local.file
	_ "github.com/grafana/agent/component/loki/process"                         // Import loki.process
	_ "github.com/grafana/agent/component/loki/relabel"                         // Import loki.relabel
	_ "github.com/grafana/agent/component/loki/source/cloudflare"               // Import loki.source.cloudflare
	_ "github.com/grafana/agent/component/loki/source/docker"                   // Import loki.source.docker
	_ "github.com/grafana/agent/component/loki/source/file"                     // Import loki.source.file
	_ "github.com/grafana/agent/component/loki/source/gcplog"                   // Import loki.source.gcplog
	_ "github.com/grafana/agent/component/loki/source/gelf"                     // Import loki.source.gelf
	_ "github.com/grafana/agent/component/loki/source/heroku"                   // Import loki.source.heroku
	_ "github.com/grafana/agent/component/loki/source/journal"                  // Import loki.source.journal
	_ "github.com/grafana/agent/component/loki/source/kubernetes"               // Import loki.source.kubernetes
	_ "github.com/grafana/agent/component/loki/source/podlogs"                  // Import loki.source.podlogs
	_ "github.com/grafana/agent/component/loki/source/syslog"                   // Import loki.source.syslog
	_ "github.com/grafana/agent/component/loki/source/windowsevent"             // Import loki.source.windowsevent
	_ "github.com/grafana/agent/component/loki/write"                           // Import loki.write
	_ "github.com/grafana/agent/component/mimir/rules/kubernetes"               // Import mimir.rules.kubernetes
	_ "github.com/grafana/agent/component/otelcol/auth/basic"                   // Import otelcol.auth.basic
	_ "github.com/grafana/agent/component/otelcol/auth/bearer"                  // Import otelcol.auth.bearer
	_ "github.com/grafana/agent/component/otelcol/auth/headers"                 // Import otelcol.auth.headers
	_ "github.com/grafana/agent/component/otelcol/exporter/jaeger"              // Import otelcol.exporter.jaeger
	_ "github.com/grafana/agent/component/otelcol/exporter/loki"                // Import otelcol.exporter.loki
	_ "github.com/grafana/agent/component/otelcol/exporter/otlp"                // Import otelcol.exporter.otlp
	_ "github.com/grafana/agent/component/otelcol/exporter/otlphttp"            // Import otelcol.exporter.otlphttp
	_ "github.com/grafana/agent/component/otelcol/exporter/prometheus"          // Import otelcol.exporter.prometheus
	_ "github.com/grafana/agent/component/otelcol/processor/batch"              // Import otelcol.processor.batch
	_ "github.com/grafana/agent/component/otelcol/processor/memorylimiter"      // Import otelcol.processor.memory_limiter
	_ "github.com/grafana/agent/component/otelcol/processor/tail_sampling"      // Import otelcol.processor.tail_sampling
	_ "github.com/grafana/agent/component/otelcol/receiver/jaeger"              // Import otelcol.receiver.jaeger
	_ "github.com/grafana/agent/component/otelcol/receiver/kafka"               // Import otelcol.receiver.kafka
	_ "github.com/grafana/agent/component/otelcol/receiver/loki"                // Import otelcol.receiver.loki
	_ "github.com/grafana/agent/component/otelcol/receiver/opencensus"          // Import otelcol.receiver.opencensus
	_ "github.com/grafana/agent/component/otelcol/receiver/otlp"                // Import otelcol.receiver.otlp
	_ "github.com/grafana/agent/component/otelcol/receiver/prometheus"          // Import otelcol.receiver.prometheus
	_ "github.com/grafana/agent/component/otelcol/receiver/zipkin"              // Import otelcol.receiver.zipkin
	_ "github.com/grafana/agent/component/phlare/scrape"                        // Import phlare.scrape
	_ "github.com/grafana/agent/component/phlare/write"                         // Import phlare.write
	_ "github.com/grafana/agent/component/prometheus/integration/node_exporter" // Import prometheus.integration.node_exporter
	_ "github.com/grafana/agent/component/prometheus/relabel"                   // Import prometheus.relabel
	_ "github.com/grafana/agent/component/prometheus/remotewrite"               // Import prometheus.remote_write
	_ "github.com/grafana/agent/component/prometheus/scrape"                    // Import prometheus.scrape
	_ "github.com/grafana/agent/component/remote/http"                          // Import remote.http
	_ "github.com/grafana/agent/component/remote/s3"                            // Import remote.s3
)
