package prometheusconvert

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	prom_config "github.com/prometheus/prometheus/config"
)

func appendPrometheusRemoteWrite(pb *prometheusBlocks, globalConfig prom_config.GlobalConfig, remoteWriteConfigs []*prom_config.RemoteWriteConfig, label string) *remotewrite.Exports {
	remoteWriteArgs := toRemotewriteArguments(globalConfig, remoteWriteConfigs)

	remoteWriteLabel := label
	if remoteWriteLabel == "" {
		remoteWriteLabel = "default"
	}

	if len(remoteWriteConfigs) > 0 {
		name := []string{"prometheus", "remote_write"}
		block := common.NewBlockWithOverride(name, remoteWriteLabel, remoteWriteArgs)

		names := []string{}
		for _, remoteWriteConfig := range remoteWriteConfigs {
			names = append(names, remoteWriteConfig.Name)
		}
		summary := fmt.Sprintf("Converted %d remote_write[s] %q into...", len(remoteWriteConfigs), strings.Join(names, ","))
		detail := fmt.Sprintf("	A prometheus.remote_write.%s component", remoteWriteLabel)
		pb.prometheusRemoteWriteBlocks = append(pb.prometheusRemoteWriteBlocks, newPrometheusBlock(block, name, label, summary, detail))
	}

	return &remotewrite.Exports{
		Receiver: common.ConvertAppendable{Expr: "prometheus.remote_write." + remoteWriteLabel + ".receiver"},
	}
}

func validateRemoteWriteConfig(remoteWriteConfig *prom_config.RemoteWriteConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if remoteWriteConfig.SigV4Config != nil {
		diags.Add(diag.SeverityLevelError, "unsupported remote_write sigv4 config was provided")
	}

	newDiags := ValidateHttpClientConfig(&remoteWriteConfig.HTTPClientConfig)
	diags = append(diags, newDiags...)

	return diags
}

func toRemotewriteArguments(globalConfig prom_config.GlobalConfig, remoteWriteConfigs []*prom_config.RemoteWriteConfig) *remotewrite.Arguments {
	externalLabels := globalConfig.ExternalLabels.Map()
	if len(externalLabels) == 0 {
		externalLabels = nil
	}

	return &remotewrite.Arguments{
		ExternalLabels: externalLabels,
		Endpoints:      getEndpointOptions(remoteWriteConfigs),
		WALOptions:     remotewrite.DefaultWALOptions,
	}
}

func getEndpointOptions(remoteWriteConfigs []*prom_config.RemoteWriteConfig) []*remotewrite.EndpointOptions {
	endpoints := make([]*remotewrite.EndpointOptions, 0)

	for _, remoteWriteConfig := range remoteWriteConfigs {
		endpoint := &remotewrite.EndpointOptions{
			Name:                 remoteWriteConfig.Name,
			URL:                  remoteWriteConfig.URL.String(),
			RemoteTimeout:        time.Duration(remoteWriteConfig.RemoteTimeout),
			Headers:              remoteWriteConfig.Headers,
			SendExemplars:        remoteWriteConfig.SendExemplars,
			SendNativeHistograms: remoteWriteConfig.SendNativeHistograms,
			HTTPClientConfig:     ToHttpClientConfig(&remoteWriteConfig.HTTPClientConfig),
			QueueOptions:         toQueueOptions(&remoteWriteConfig.QueueConfig),
			MetadataOptions:      toMetadataOptions(&remoteWriteConfig.MetadataConfig),
			WriteRelabelConfigs:  ToFlowRelabelConfigs(remoteWriteConfig.WriteRelabelConfigs),
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func toQueueOptions(queueConfig *prom_config.QueueConfig) *remotewrite.QueueOptions {
	return &remotewrite.QueueOptions{
		Capacity:          queueConfig.Capacity,
		MaxShards:         queueConfig.MaxShards,
		MinShards:         queueConfig.MinShards,
		MaxSamplesPerSend: queueConfig.MaxSamplesPerSend,
		BatchSendDeadline: time.Duration(queueConfig.BatchSendDeadline),
		MinBackoff:        time.Duration(queueConfig.MinBackoff),
		MaxBackoff:        time.Duration(queueConfig.MaxBackoff),
		RetryOnHTTP429:    queueConfig.RetryOnRateLimit,
	}
}

func toMetadataOptions(metadataConfig *prom_config.MetadataConfig) *remotewrite.MetadataOptions {
	return &remotewrite.MetadataOptions{
		Send:              metadataConfig.Send,
		SendInterval:      time.Duration(metadataConfig.SendInterval),
		MaxSamplesPerSend: metadataConfig.MaxSamplesPerSend,
	}
}
