package prometheusconvert

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/river/rivertypes"
	"github.com/prometheus/common/sigv4"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/storage/remote/azuread"
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
		pb.prometheusRemoteWriteBlocks = append(pb.prometheusRemoteWriteBlocks, newPrometheusBlock(block, name, remoteWriteLabel, summary, detail))
	}

	return &remotewrite.Exports{
		Receiver: common.ConvertAppendable{Expr: "prometheus.remote_write." + remoteWriteLabel + ".receiver"},
	}
}

func validateRemoteWriteConfig(remoteWriteConfig *prom_config.RemoteWriteConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddAll(ValidateHttpClientConfig(&remoteWriteConfig.HTTPClientConfig))
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
			SigV4:                toSigV4(remoteWriteConfig.SigV4Config),
			AzureAD:              toAzureAD(remoteWriteConfig.AzureADConfig),
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

// toSigV4 converts a Prometheus SigV4 config to a River SigV4 config.
func toSigV4(sigv4Config *sigv4.SigV4Config) *remotewrite.SigV4Config {
	if sigv4Config == nil {
		return nil
	}

	return &remotewrite.SigV4Config{
		Region:    sigv4Config.Region,
		AccessKey: sigv4Config.AccessKey,
		SecretKey: rivertypes.Secret(sigv4Config.SecretKey),
		Profile:   sigv4Config.Profile,
		RoleARN:   sigv4Config.RoleARN,
	}
}

// toAzureAD converts a Prometheus AzureAD config to a River AzureAD config.
func toAzureAD(azureADConfig *azuread.AzureADConfig) *remotewrite.AzureADConfig {
	if azureADConfig == nil {
		return nil
	}

	return &remotewrite.AzureADConfig{
		Cloud: azureADConfig.Cloud,
		ManagedIdentity: remotewrite.ManagedIdentityConfig{
			ClientID: azureADConfig.ManagedIdentity.ClientID,
		},
	}
}
