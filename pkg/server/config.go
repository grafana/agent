package server

import (
	"flag"

	"github.com/weaveworks/common/logging"
)

// Config holds dynamic configuration options for a Server.
type Config struct {
	LogLevel  logging.Level  `yaml:"log_level"`
	LogFormat logging.Format `yaml:"log_format"`

	GRPC GRPCConfig `yaml:",inline"`
	HTTP HTTPConfig `yaml:",inline"`

	// Flags is a DEPRECATED field holding static coniguration options.
	// It will be removed from YAML and only be exposed by command-line flags in
	// v0.26.0.
	//
	// Updating any field found in Flags will cause updating the Server to fail.
	Flags Flags `yaml:",inline"`
}

// UnmarshalYAML unmarshals the server config with defaults applied.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig

	type config Config
	return unmarshal((*config)(c))
}

// HTTPConfig holds dynamic configuration options for the HTTP server.
type HTTPConfig struct {
	TLSConfig TLSConfig `yaml:"http_tls_config"`
}

// GRPCConfig holds dynamic configuration options for the gRPC server.
type GRPCConfig struct {
	TLSConfig TLSConfig `yaml:"grpc_tls_config"`
}

// Default configuration structs.
var (
	DefaultConfig = Config{
		GRPC:      DefaultGRPCConfig,
		HTTP:      DefaultHTTPConfig,
		Flags:     DefaultFlags,
		LogLevel:  DefaultLogLevel,
		LogFormat: DefaultLogFormat,
	}

	DefaultHTTPConfig = HTTPConfig{
		// No non-zero defaults yet
	}

	DefaultGRPCConfig = GRPCConfig{
		// No non-zero defaults yet
	}

	emptyFlagSet    = flag.NewFlagSet("", flag.ExitOnError)
	DefaultLogLevel = func() logging.Level {
		var lvl logging.Level
		lvl.RegisterFlags(emptyFlagSet)
		return lvl
	}()
	DefaultLogFormat = func() logging.Format {
		var fmt logging.Format
		fmt.RegisterFlags(emptyFlagSet)
		return fmt
	}()
)

// RegisterFlags registers flags for c to the given FlagSet.
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	c.Flags.RegisterFlags(f)
}
