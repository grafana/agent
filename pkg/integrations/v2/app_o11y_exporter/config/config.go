package config

const (
	DEFAULT_RATE_LIMITING_RPS       = 100
	DEFAULT_RATE_LIMITING_BURSTINES = 50
	DEFAULT_MAX_PAYLOAD_SIZE        = 5e6
)

var DefaultConfig = AppExporterConfig{
	// Default JS agent port
	CORSAllowedOrigins: []string{"http://localhost:1234"},
	RateLimiting: RateLimiting{
		Enabled:    false,
		RPS:        DEFAULT_RATE_LIMITING_RPS,
		Burstiness: DEFAULT_RATE_LIMITING_BURSTINES,
	},
	MaxAllowedPayloadSize: DEFAULT_MAX_PAYLOAD_SIZE,
	SourceMap: SourceMapConfig{
		Enabled: false,
		MapURI:  "",
	},
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	},
	LogsInstance: "default",
	Measurements: []Measurement{},
}

type ServerConfig struct {
	Host string `yaml:"host,omitempty"`
	Port int    `yaml:"port,omitempty"`
}

type RateLimiting struct {
	Enabled    bool    `yaml:"enabled,omitempty"`
	RPS        float64 `yaml:"rps,omitempty"`
	Burstiness int     `yaml:"burstiness,omitempty"`
}

type SourceMapConfig struct {
	Enabled bool   `yaml:"enabled,omitempty"`
	MapURI  string `yaml:"map_uri,omitempty"`
}

type Measurement struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type AppExporterConfig struct {
	CORSAllowedOrigins    []string        `yaml:"cors_allowed_origins,omitempty"`
	RateLimiting          RateLimiting    `yaml:"rate_limiting,omitempty"`
	MaxAllowedPayloadSize int64           `yaml:"max_allowed_payload_size,omitempty"`
	SourceMap             SourceMapConfig `yaml:"source_map,omitempty"`
	Server                ServerConfig    `yaml:"server,omitempty"`
	LogsInstance          string          `yaml:"logs_instance"`
	Measurements          []Measurement   `yaml:"custom_measurements"`
}

func (c *AppExporterConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig
	type cA AppExporterConfig

	if err := unmarshal((*cA)(c)); err != nil {
		return err
	}

	if c.RateLimiting.Enabled && c.RateLimiting.RPS == 0 {
		c.RateLimiting.RPS = DEFAULT_RATE_LIMITING_RPS
	}

	return nil
}
