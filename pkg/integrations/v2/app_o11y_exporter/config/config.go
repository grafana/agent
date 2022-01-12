package config

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
