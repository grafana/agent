// Package all imports all known component packages.
package all

import (
	_ "github.com/grafana/agent/component/local/file"           // Import local.file
	_ "github.com/grafana/agent/component/otel/jaegerreceiver"  // Import otel.receiver_jaeger
	_ "github.com/grafana/agent/component/otel/loggingexporter" // Import otel.exporter_logging
	_ "github.com/grafana/agent/component/otel/otlpexporter"    // Import otel.receiver_otlp
	_ "github.com/grafana/agent/component/otel/otlpreceiver"    // Import otel.exporter_otlp
	_ "github.com/grafana/agent/component/otel/zipkinreceiver"  // Import otel.receiver_zipkin
	_ "github.com/grafana/agent/component/targets/mutate"       // Import targets.mutate
)
