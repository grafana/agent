package vmware_exporter

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/common"
	"github.com/grafana/agent/pkg/integrations/v2/metricsutils"
	"github.com/grafana/vmware_exporter/vsphere"
)

func init() {
	integrations.Register(&Config{}, integrations.TypeMultiplex)
}

var DefaultConfig = Config{
	ChunkSize:               256,
	CollectConcurrency:      8,
	ObjectDiscoveryInterval: 0,
	EnableExporterMetrics:   true,
}

type Config struct {
	ChunkSize               int                  `yaml:"chunk_size,omitempty"`
	CollectConcurrency      int                  `yaml:"collect_concurrency,omitempty"`
	VSphereURL              string               `yaml:"vsphere_url,omitempty"`
	VSphereUser             string               `yaml:"vsphere_user,omitempty"`
	VSpherePass             string               `yaml:"vsphere_password,omitempty"`
	ObjectDiscoveryInterval time.Duration        `yaml:"discovery_interval,omitempty"`
	EnableExporterMetrics   bool                 `yaml:"enable_exporter_metrics,omitempty"`
	Common                  common.MetricsConfig `yaml:",inline"`
}

var _ integrations.Config = (*Config)(nil)

// UnmarshalYAML implements the Unmarshaler interface
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig
	type plain Config
	return unmarshal((*plain)(c))
}

func (c *Config) Name() string {
	return "vsphere"
}

func (c *Config) ApplyDefaults(g integrations.Globals) error {
	c.Common.ApplyDefaults(g.SubsystemOpts.Metrics.Autoscrape)
	return nil
}

func (c *Config) Identifier(g integrations.Globals) (string, error) {
	return *c.Common.InstanceKey, nil
}

// InstanceKey returns the vsphere instance
func (c *Config) InstanceKey(agentKey string) (string, error) {
	u, err := url.Parse(c.VSphereURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", u.Hostname(), u.Port()), nil
}

func (c *Config) NewIntegration(log log.Logger, g integrations.Globals) (integrations.Integration, error) {
	vsphereURL, err := url.Parse(c.VSphereURL)
	if err != nil {
		return nil, err
	}
	vsphereURL.User = url.UserPassword(c.VSphereUser, c.VSpherePass)

	exporterConfig := vsphere.Config{
		TelemetryPath:           "/integrations/vsphere/metrics",
		ChunkSize:               c.ChunkSize,
		CollectConcurrency:      c.CollectConcurrency,
		VSphereURL:              vsphereURL,
		ObjectDiscoveryInterval: c.ObjectDiscoveryInterval,
		EnableExporterMetrics:   c.EnableExporterMetrics,
	}
	exporter, err := vsphere.NewExporter(log, &exporterConfig)
	if err != nil {
		return nil, err
	}

	return metricsutils.NewMetricsHandlerIntegration(
		log, c, c.Common, g, exporter,
	)
}
