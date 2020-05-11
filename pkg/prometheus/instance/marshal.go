package instance

import (
	"io"

	"gopkg.in/yaml.v2"
)

// UnmarshalConfig unmarshals an instance config from a reader based on a
// provided content type.
func UnmarshalConfig(r io.Reader) (*Config, error) {
	var cfg Config
	err := yaml.NewDecoder(r).Decode(&cfg)
	return &cfg, err
}

// MarshalConfig marshals an instance config based on a provided content type.
func MarshalConfig(c *Config, sanitize bool) (string, error) {
	if sanitize {
		for _, rw := range c.RemoteWrite {
			rw.Sanitize()
		}
	}

	bb, err := yaml.Marshal(c)
	return string(bb), err
}
