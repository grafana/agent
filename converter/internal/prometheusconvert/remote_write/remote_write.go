package remote_write

import (
	"time"

	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/agent/converter/internal/prometheusconvert/common"
	promconfig "github.com/prometheus/prometheus/config"
)

func Reconvert(remoteWriteConfigs []*promconfig.RemoteWriteConfig) *remotewrite.Arguments {
	return &remotewrite.Arguments{
		ExternalLabels: map[string]string{},
		Endpoints:      getEndpointOptions(remoteWriteConfigs),
		WALOptions:     remotewrite.WALOptions{},
	}
}

func getEndpointOptions(remoteWriteConfigs []*promconfig.RemoteWriteConfig) []*remotewrite.EndpointOptions {
	endpoints := make([]*remotewrite.EndpointOptions, 0)

	for _, remoteWriteConfig := range remoteWriteConfigs {
		endpoint := &remotewrite.EndpointOptions{
			Name:                 remoteWriteConfig.Name,
			URL:                  remoteWriteConfig.URL.String(),
			RemoteTimeout:        time.Duration(remoteWriteConfig.RemoteTimeout),
			Headers:              remoteWriteConfig.Headers,
			SendExemplars:        remoteWriteConfig.SendExemplars,
			SendNativeHistograms: remoteWriteConfig.SendNativeHistograms,
			HTTPClientConfig:     common.ReconvertHttpClientConfig(&remoteWriteConfig.HTTPClientConfig),
			QueueOptions:         reconvertQueueConfig(&remoteWriteConfig.QueueConfig),
			MetadataOptions:      reconvertMetadataConfig(&remoteWriteConfig.MetadataConfig),
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func reconvertQueueConfig(queueConfig *promconfig.QueueConfig) *remotewrite.QueueOptions {
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

func reconvertMetadataConfig(metadataConfig *promconfig.MetadataConfig) *remotewrite.MetadataOptions {
	return &remotewrite.MetadataOptions{
		Send:              metadataConfig.Send,
		SendInterval:      time.Duration(metadataConfig.SendInterval),
		MaxSamplesPerSend: metadataConfig.MaxSamplesPerSend,
	}
}
