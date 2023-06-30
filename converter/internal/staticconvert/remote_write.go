package staticconvert

import (
	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/prometheus/prometheus/config"
)

func appendPrometheusRemoteWrite(pb *blocks, globalRemoteWriteConfig []*config.RemoteWriteConfig, instanceConfig instance.Config, label string) *remotewrite.Exports {
	remoteWriteArgs := toRemotewriteArguments(globalRemoteWriteConfig, instanceConfig)
	block := common.NewBlockWithOverride([]string{"prometheus", "remote_write"}, label, remoteWriteArgs)
	pb.prometheusRemoteWriteBlocks = append(pb.prometheusRemoteWriteBlocks, block)
	return &remotewrite.Exports{
		Receiver: common.ConvertAppendable{Expr: "prometheus.remote_write." + label + ".receiver"},
	}
}

func toRemotewriteArguments(globalRemoteWriteConfig []*config.RemoteWriteConfig, instanceConfig instance.Config) *remotewrite.Arguments {
	var endpoints []*remotewrite.EndpointOptions
	if len(instanceConfig.RemoteWrite) == 0 {
		// use the global remote write if we don't have an instance set
		endpoints = prometheusconvert.GetEndpointOptions(globalRemoteWriteConfig)
	} else {
		endpoints = prometheusconvert.GetEndpointOptions(instanceConfig.RemoteWrite)
	}

	return &remotewrite.Arguments{
		ExternalLabels: nil,
		Endpoints:      endpoints,
		WALOptions:     toWALOptions(instanceConfig),
	}
}

func toWALOptions(instanceConfig instance.Config) remotewrite.WALOptions {
	return remotewrite.WALOptions{
		TruncateFrequency: instanceConfig.WALTruncateFrequency,
		MinKeepaliveTime:  instanceConfig.MinWALTime,
		MaxKeepaliveTime:  instanceConfig.MaxWALTime,
	}
}
