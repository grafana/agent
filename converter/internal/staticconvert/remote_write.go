package staticconvert

import (
	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/metrics"
	"github.com/grafana/agent/pkg/metrics/instance"
)

func appendPrometheusRemoteWrite(pb *blocks, metricsConfig *metrics.Config, instanceConfig instance.Config, label string) *remotewrite.Exports {
	remoteWriteArgs := toRemotewriteArguments(metricsConfig, instanceConfig)
	block := common.NewBlockWithOverride([]string{"prometheus", "remote_write"}, label, remoteWriteArgs)
	pb.prometheusRemoteWriteBlocks = append(pb.prometheusRemoteWriteBlocks, block)
	return &remotewrite.Exports{
		Receiver: common.ConvertAppendable{Expr: "prometheus.remote_write." + label + ".receiver"},
	}
}

func toRemotewriteArguments(metricsConfig *metrics.Config, instanceConfig instance.Config) *remotewrite.Arguments {
	var endpoints []*remotewrite.EndpointOptions
	if len(instanceConfig.RemoteWrite) == 0 {
		// use the global remote write if we don't have an instance set
		endpoints = prometheusconvert.GetEndpointOptions(metricsConfig.Global.RemoteWrite)
	} else {
		endpoints = prometheusconvert.GetEndpointOptions(instanceConfig.RemoteWrite)
	}

	return &remotewrite.Arguments{
		ExternalLabels: nil,
		Endpoints:      endpoints,
		WALOptions:     toWALOptions(metricsConfig),
	}
}

func toWALOptions(metricsConfig *metrics.Config) remotewrite.WALOptions {
	return remotewrite.WALOptions{
		TruncateFrequency: metricsConfig.WALCleanupPeriod,
		MinKeepaliveTime:  remotewrite.DefaultWALOptions.MinKeepaliveTime,
		MaxKeepaliveTime:  metricsConfig.WALCleanupAge,
	}
}
