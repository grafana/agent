package frontend_receiver //nolint:golint

import (
	"fmt"
	"time"

	"github.com/grafana/agent/pkg/integrations/config"
	loki "github.com/grafana/agent/pkg/logs"
	"github.com/grafana/agent/pkg/tempo"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/integrations"
	recconf "github.com/grafana/grafana-frontend-telemetry-receiver/pkg/config"
	"github.com/grafana/grafana-frontend-telemetry-receiver/pkg/utils"
	prommodel "github.com/prometheus/common/model"
)

// Config controls the frontend_receiver integration.
type Config struct {
	Common           config.Common          `mapstructure:",inline" yaml:",inline"`
	Receiver         recconf.ReceiverConfig `mapstructure:",inline" yaml:",inline"`
	Endpoint         string                 `mapstructure:"endpoint" yaml:"endpoint"`
	LogsInstanceName string                 `mapstructure:"logs_instance_name" yaml:"logs_instance_name"`
	LogsTimeout      time.Duration          `mapstructure:"logs_timeout" yaml:"logs_timeout,omitempty"`
	LogsLabels       map[string]string      `mapstructure:"logs_labels" yaml:"logs_labels"`
}

// CommonConfig returns the set of common settings shared across all integrations.
func (c *Config) CommonConfig() config.Common {
	return c.Common
}

// Name returns the name of the integration that this config is for.
func (c *Config) Name() string {
	return "frontend_receiver"
}

func (c *Config) labelSet(kv *utils.KeyVal) prommodel.LabelSet {
	set := make(prommodel.LabelSet)
	for key, value := range c.LogsLabels {
		if len(value) > 0 {
			set[prommodel.LabelName(key)] = prommodel.LabelValue(value)
		} else {
			val, ok := kv.Get(key)
			if ok == true {
				set[prommodel.LabelName(key)] = prommodel.LabelValue(fmt.Sprint(val))
			}
		}
	}
	return set
}

func (c *Config) lokiInstance(logs *loki.Logs) (*loki.Instance, error) {
	if len(c.LogsInstanceName) > 1 {
		instance := logs.Instance(c.LogsInstanceName)
		if instance == nil {
			return nil, fmt.Errorf("frontend receiver references loki instance \"%s\", but no such loki instance is configured", c.LogsInstanceName)
		}
		return instance, nil
	}
	return nil, nil
}

// NewIntegration converts this config into an instance of an integration.
func (c *Config) NewIntegration(l log.Logger, loki *loki.Logs, tempo *tempo.Tempo) (integrations.Integration, error) {
	lokiInstance, err := c.lokiInstance(loki)
	if err != nil {
		return nil, err
	}
	return New(l, c, lokiInstance)
}
