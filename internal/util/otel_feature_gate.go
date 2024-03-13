package util

import (
	"fmt"

	"go.opentelemetry.io/collector/featuregate"
	_ "go.opentelemetry.io/collector/obsreport"
)

// Enables a set of feature gates in Otel's Global Feature Gate Registry.
func EnableOtelFeatureGates(fgNames ...string) error {
	fgReg := featuregate.GlobalRegistry()

	for _, fg := range fgNames {
		err := fgReg.Set(fg, true)
		if err != nil {
			return fmt.Errorf("error setting Otel feature gate: %w", err)
		}
	}

	return nil
}

var (
	// disableHighCardinalityMetrics is required so that we don't include labels containing ports and IP addresses in gRPC metrics.
	// Example metric with high cardinality...
	// rpc_server_duration_bucket{net_sock_peer_addr="127.0.0.1",net_sock_peer_port="59947",rpc_grpc_status_code="0",rpc_method="Export",rpc_service="opentelemetry.proto.collector.trace.v1.TraceService",rpc_system="grpc",traces_config="default",le="7500"} 294
	// ... the same metric when disableHighCardinalityMetrics is switched on looks like this:
	// rpc_server_duration_bucket{rpc_grpc_status_code="0",rpc_method="Export",rpc_service="opentelemetry.proto.collector.trace.v1.TraceService",rpc_system="grpc",traces_config="default",le="7500"} 32
	// For more context:
	// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/rpc-metrics/
	// https://github.com/open-telemetry/opentelemetry-go-contrib/pull/2700
	// https://github.com/open-telemetry/opentelemetry-collector/pull/6788/files
	staticModeOtelFeatureGates = []string{
		"telemetry.disableHighCardinalityMetrics",
	}

	flowModeOtelFeatureGates = []string{}
)

// Enables a set of feature gates which should always be enabled for Static mode.
func SetupStaticModeOtelFeatureGates() error {
	return EnableOtelFeatureGates(staticModeOtelFeatureGates...)
}

// Enables a set of feature gates which should always be enabled for Flow mode.
func SetupFlowModeOtelFeatureGates() error {
	return EnableOtelFeatureGates(flowModeOtelFeatureGates...)
}
