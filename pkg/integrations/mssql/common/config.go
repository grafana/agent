package common

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// DefaultConfig is the default config for the mssql integration
var DefaultConfig = Config{
	MaxIdleConnections: 3,
	MaxOpenConnections: 3,
	Timeout:            10 * time.Second,
}

// Config is the configuration for the mssql integration
type Config struct {
	ConnectionString   string        `yaml:"connection_string,omitempty"`
	MaxIdleConnections int           `yaml:"max_idle_connections,omitempty"`
	MaxOpenConnections int           `yaml:"max_open_connections,omitempty"`
	Timeout            time.Duration `yaml:"timeout,omitempty"`
}

func (c Config) Validate() error {
	if c.ConnectionString == "" {
		return errors.New("the connection_string parameter is required")
	}

	url, err := url.Parse(c.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection_string: %w", err)
	}

	if url.Scheme != "sqlserver" {
		return errors.New("scheme of provided connection_string URL must be sqlserver")
	}

	if c.MaxOpenConnections < 1 {
		return errors.New("max_connections must be at least 1")
	}

	if c.MaxIdleConnections < 1 {
		return errors.New("max_idle_connection must be at least 1")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type plain Config
	return unmarshal((*plain)(c))
}
