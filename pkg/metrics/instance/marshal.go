package instance

import (
	"bytes"
	"io"

	config_util "github.com/prometheus/common/config"
	"gopkg.in/yaml.v2"
)

// UnmarshalConfig unmarshals an instance shared from a reader based on a
// provided content type.
func UnmarshalConfig(r io.Reader) (*Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)
	err := dec.Decode(&cfg)
	return &cfg, err
}

// MarshalConfig marshals an instance shared based on a provided content type.
func MarshalConfig(c *Config, scrubSecrets bool) ([]byte, error) {
	var buf bytes.Buffer
	err := MarshalConfigToWriter(c, &buf, scrubSecrets)
	return buf.Bytes(), err
}

// MarshalConfigToWriter marshals a shared to an io.Writer.
func MarshalConfigToWriter(c *Config, w io.Writer, scrubSecrets bool) error {
	enc := yaml.NewEncoder(w)

	// If we're not sanitizing the marshaled shared, we want to add in an
	// encoding hook to ignore how Secrets marshal (i.e., scrubbing the value
	// and replacing it with <secret>).
	if !scrubSecrets {
		enc.SetHook(func(in interface{}) (ok bool, out interface{}, err error) {
			switch v := in.(type) {
			case config_util.Secret:
				return true, string(v), nil
			default:
				return false, nil, nil
			}
		})
	}

	type plain Config
	return enc.Encode((*plain)(c))
}
