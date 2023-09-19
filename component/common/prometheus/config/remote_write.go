package config

import (
	"time"

	common "github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/common/config/sigv4config"
	"github.com/grafana/agent/component/common/prometheus/storage/remote/azuread"
	"github.com/grafana/agent/component/common/relabel"
	"github.com/prometheus/common/model"
	internal "github.com/prometheus/prometheus/config"
)

type RemoteWriteConfig struct {
	uRL                  common.URL        `river:"url,attr"`
	remoteTimeout        time.Duration     `river:"remote_timeout,attr,optional"`
	headers              map[string]string `river:"headers,attr,optional"`
	writeRelabelConfigs  []*relabel.Config `river:"write_relabel_configs,block,optional"`
	name                 string            `river:"name,string,optional"`
	sendExamplars        bool              `river:"send_examplars,bool,optional"`
	sendNativeHistograms bool              `river:"send_native_histograms,bool,optional`

	httpClientConfig common.HTTPClientConfig `river:",squash"`
	queueConfig      QueueConfig             `river:"queue_config,block,optional"`
	metadataConfig   MetadataConfig          `river:"metadata_config,block,optional"`
	sigV4Config      sigv4config.SigV4Config `river:"sigv4,block,optional"`
	azureADConfig    azuread.AzureADConfig   `river:"azure_ad_config,block,optional"`
}

type QueueConfig struct {
	Capacity          int           `river:"capacity,number,optional"`
	MaxShards         int           `river:"max_shards,number,optional"`
	MinShards         int           `river:"min_shards,number,optional"`
	MaxSamplesPerSend int           `river:"max_samples_per_send,number,optional"`
	BatchSendDeadline time.Duration `river:"batch_send_deadline,attr,optional"`
	MinBackoff        time.Duration `river:"min_backoff,attr,optional"`
	MaxBackoff        time.Duration `river:"max_backoff,attr,optional"`
	RetryOnRateLimit  bool          `river:"retry_on_rate_limit,bool,optional"`
}

func (c *QueueConfig) ToInternal() internal.QueueConfig {
	return internal.QueueConfig{
		Capacity:          c.Capacity,
		MaxShards:         c.MaxShards,
		MinShards:         c.MinShards,
		MaxSamplesPerSend: c.MaxSamplesPerSend,
		BatchSendDeadline: model.Duration(c.BatchSendDeadline),
		MinBackoff:        model.Duration(c.MinBackoff),
		MaxBackoff:        model.Duration(c.MaxBackoff),
		RetryOnRateLimit:  c.RetryOnRateLimit,
	}
}

type MetadataConfig struct {
	Send              bool          `river:"send,bool,optional"`
	SendInterval      time.Duration `river:"send_interval,attr,optional"`
	MaxSamplesPerSend int           `river:"max_samples_per_send,number,optional"`
}

func (c *MetadataConfig) ToInternal() internal.MetadataConfig {
	return internal.MetadataConfig{
		Send:              c.Send,
		SendInterval:      model.Duration(c.SendInterval),
		MaxSamplesPerSend: c.MaxSamplesPerSend,
	}
}

func (c *RemoteWriteConfig) ToInternal() internal.RemoteWriteConfig {
	url := c.uRL.Convert()
	sigV4Config := c.sigV4Config.ToInternal()
	azureADConfig := c.azureADConfig.ToInternal()
	return internal.RemoteWriteConfig{
		URL:                  &url,
		RemoteTimeout:        model.Duration(c.remoteTimeout),
		Headers:              c.headers,
		WriteRelabelConfigs:  relabel.ComponentToPromRelabelConfigs(c.writeRelabelConfigs),
		Name:                 c.name,
		SendExemplars:        c.sendExamplars,
		SendNativeHistograms: c.sendNativeHistograms,
		HTTPClientConfig:     *c.httpClientConfig.Convert(),
		QueueConfig:          c.queueConfig.ToInternal(),
		MetadataConfig:       c.metadataConfig.ToInternal(),
		SigV4Config:          &sigV4Config,
		AzureADConfig:        &azureADConfig,
	}
}

func RemoteWriteConfigsToInternals(cs []*RemoteWriteConfig) []*internal.RemoteWriteConfig {
	var ret []*internal.RemoteWriteConfig

	for _, c := range cs {
		internal := c.ToInternal()
		ret = append(ret, &internal)
	}
	return ret
}
