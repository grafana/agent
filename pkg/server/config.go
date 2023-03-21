package server

import (
	"flag"

	"github.com/weaveworks/common/logging"
)

// LogLevel wraps the logging.Level type to allow defining IsZero, which is required to make omitempty work when marshalling YAML.
type LogLevel struct {
	logging.Level `yaml:",inline"`
}

func (l LogLevel) IsZero() bool {
	return l.Level.String() == ""
}

// Config holds dynamic configuration options for a Server.
type Config struct {
	LogLevel  LogLevel       `yaml:"log_level,omitempty"`
	LogFormat logging.Format `yaml:"log_format,omitempty"`

	GRPC GRPCConfig `yaml:",inline"`
	HTTP HTTPConfig `yaml:",inline"`
}

// UnmarshalYAML unmarshals the server config with defaults applied.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig()

	type config Config
	return unmarshal((*config)(c))
}

// HTTPConfig holds dynamic configuration options for the HTTP server.
type HTTPConfig struct {
	TLSConfig TLSConfig `yaml:"http_tls_config,omitempty"`
}

// GRPCConfig holds dynamic configuration options for the gRPC server.
type GRPCConfig struct {
	TLSConfig TLSConfig `yaml:"grpc_tls_config,omitempty"`
}

// Default configuration structs.
var (
	emptyFlagSet    = flag.NewFlagSet("", flag.ExitOnError)
	DefaultLogLevel = func() LogLevel {
		var lvl LogLevel
		lvl.RegisterFlags(emptyFlagSet)
		return lvl
	}()
	DefaultLogFormat = func() logging.Format {
		var fmt logging.Format
		fmt.RegisterFlags(emptyFlagSet)
		return fmt
	}()
)

func DefaultConfig() Config {
	DefaultHTTPConfig := HTTPConfig{
		// No non-zero defaults yet
	}

	DefaultGRPCConfig := GRPCConfig{
		// No non-zero defaults yet
	}

	return Config{
		GRPC:      DefaultGRPCConfig,
		HTTP:      DefaultHTTPConfig,
		LogLevel:  DefaultLogLevel,
		LogFormat: DefaultLogFormat,
	}
}
