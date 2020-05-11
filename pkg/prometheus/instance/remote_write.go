package instance

import (
	"errors"
	"fmt"
	"time"

	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/pkg/relabel"
)

var (
	DefaultRemoteWriteConfig = RemoteWriteConfig{
		RemoteTimeout:  model.Duration(30 * time.Second),
		QueueConfig:    config.DefaultQueueConfig,
		MetadataConfig: config.DefaultMetadataConfig,
	}
)

// RemoteWriteConfig holds configuration for remote_write. It is duplicated from the
// Prometheus RemoteWriteConfig to allow for marshaling secrets to YAML.
type RemoteWriteConfig struct {
	URL                 *config_util.URL  `yaml:"url"`
	RemoteTimeout       model.Duration    `yaml:"remote_timeout,omitempty"`
	WriteRelabelConfigs []*relabel.Config `yaml:"write_relabel_configs,omitempty"`
	Name                string            `yaml:"name,omitempty"`

	// We cannot do proper Go type embedding below as the parser will then parse
	// values arbitrarily into the overflow maps of further-down types.
	HTTPClientConfig HTTPClientConfig      `yaml:",inline"`
	QueueConfig      config.QueueConfig    `yaml:"queue_config,omitempty"`
	MetadataConfig   config.MetadataConfig `yaml:"metadata_config,omitempty"`
}

// PrometheusConfig returns the Prometheus-config equivalent of RemoteWriteConfig.
func (c *RemoteWriteConfig) PrometheusConfig() *config.RemoteWriteConfig {
	return &config.RemoteWriteConfig{
		URL:                 c.URL,
		RemoteTimeout:       c.RemoteTimeout,
		WriteRelabelConfigs: c.WriteRelabelConfigs,
		Name:                c.Name,

		HTTPClientConfig: c.HTTPClientConfig.PrometheusConfig(),
		QueueConfig:      c.QueueConfig,
		MetadataConfig:   c.MetadataConfig,
	}
}

// Sanitize finds all secrets in the RemoteWriteConfig and replaces their values
// with <secret>.
func (c *RemoteWriteConfig) Sanitize() {
	if c.HTTPClientConfig.BasicAuth != nil && c.HTTPClientConfig.BasicAuth.Password != "" {
		c.HTTPClientConfig.BasicAuth.Password = "<secret>"
	}

	if c.HTTPClientConfig.BearerToken != "" {
		c.HTTPClientConfig.BearerToken = "<secret>"
	}
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *RemoteWriteConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultRemoteWriteConfig
	type plain RemoteWriteConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	if c.URL == nil {
		return errors.New("url for remote_write is empty")
	}
	for _, rlcfg := range c.WriteRelabelConfigs {
		if rlcfg == nil {
			return errors.New("empty or null relabeling rule in remote write config")
		}
	}

	// The UnmarshalYAML method of HTTPClientConfig is not being called because it's not a pointer.
	// We cannot make it a pointer as the parser panics for inlined pointer structs.
	// Thus we just do its validation here.
	return c.HTTPClientConfig.Validate()
}

// HTTPClientConfig configures an HTTP client.
type HTTPClientConfig struct {
	// The HTTP basic authentication credentials for the targets.
	BasicAuth *BasicAuth `yaml:"basic_auth,omitempty"`
	// The bearer token for the targets.
	BearerToken string `yaml:"bearer_token,omitempty"`
	// The bearer token file for the targets.
	BearerTokenFile string `yaml:"bearer_token_file,omitempty"`
	// HTTP proxy server to use to connect to the targets.
	ProxyURL config_util.URL `yaml:"proxy_url,omitempty"`
	// TLSConfig to use to connect to the targets.
	TLSConfig config_util.TLSConfig `yaml:"tls_config,omitempty"`
}

// PrometheusConfig returns the Prometheus-config equivalent of HTTPClientConfig.
func (c *HTTPClientConfig) PrometheusConfig() config_util.HTTPClientConfig {
	var basic *config_util.BasicAuth
	if c.BasicAuth != nil {
		basic = &config_util.BasicAuth{
			Username: c.BasicAuth.Username,
			Password: config_util.Secret(c.BasicAuth.Password),
		}
	}

	return config_util.HTTPClientConfig{
		BasicAuth:       basic,
		BearerToken:     config_util.Secret(c.BearerToken),
		BearerTokenFile: c.BearerTokenFile,
		ProxyURL:        c.ProxyURL,
		TLSConfig:       c.TLSConfig,
	}
}

type BasicAuth struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password,omitempty"`
	PasswordFile string `yaml:"password_file,omitempty"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (c *HTTPClientConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain HTTPClientConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return c.Validate()
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (a *BasicAuth) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain BasicAuth
	return unmarshal((*plain)(a))
}

// Validate validates the HTTPClientConfig to check only one of BearerToken,
// BasicAuth and BearerTokenFile is configured.
func (c *HTTPClientConfig) Validate() error {
	if len(c.BearerToken) > 0 && len(c.BearerTokenFile) > 0 {
		return fmt.Errorf("at most one of bearer_token & bearer_token_file must be configured")
	}
	if c.BasicAuth != nil && (len(c.BearerToken) > 0 || len(c.BearerTokenFile) > 0) {
		return fmt.Errorf("at most one of basic_auth, bearer_token & bearer_token_file must be configured")
	}
	if c.BasicAuth != nil && (string(c.BasicAuth.Password) != "" && c.BasicAuth.PasswordFile != "") {
		return fmt.Errorf("at most one of basic_auth password & password_file must be configured")
	}
	return nil
}
