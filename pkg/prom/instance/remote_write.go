package instance

import (
	"fmt"

	"github.com/prometheus/prometheus/config"
)

// RemoteWriteConfig extends the default RemoteWriteConfig with extra settings.
type RemoteWriteConfig struct {
	Base config.RemoteWriteConfig `yaml:",inline"`

	SigV4 SigV4Config `yaml:"sigv4,omitempty"`
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *RemoteWriteConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	c.Base = config.DefaultRemoteWriteConfig

	type plain RemoteWriteConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	// NOTE(rfratto): UnmarshalYAML can't be called for inlined fields, which
	// must not be pointers. We need to copy the validation logic here and
	// sync changes when they happen upstream.
	if c.Base.URL == nil {
		return fmt.Errorf("url for remote_write is empty")
	}
	for _, cfg := range c.Base.WriteRelabelConfigs {
		if cfg == nil {
			return fmt.Errorf("empty or null relabeling rule in remote write config")
		}
	}

	return c.Validate()
}

// Validate validates the HTTPClientConfig along with the SigV4Config to ensure only one
// authentication mechanism is used.
func (c *RemoteWriteConfig) Validate() error {
	clientConfig := c.Base.HTTPClientConfig
	if err := clientConfig.Validate(); err != nil {
		return err
	}

	// count converts a true value to 1, allowing to sum truth conditions
	// together to calculate how many conditions were true.
	count := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}

	// Ensure at most one auth mechanism is enabled
	var (
		usingBearer = count(len(clientConfig.BearerToken) > 0 || len(clientConfig.BearerTokenFile) > 0)
		usingBasic  = count(clientConfig.BasicAuth != nil)
		usingSigV4  = count(c.SigV4.Enabled)

		enabled = usingBearer + usingBasic + usingSigV4
	)
	if enabled > 1 {
		return fmt.Errorf("at most one of sigv4, basic auth, bearer tokens must be configured")
	}

	return nil
}
