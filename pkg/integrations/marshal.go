package integrations

import (
	"bytes"
	"io"

	config_util "github.com/prometheus/common/config"
	"gopkg.in/yaml.v2"
)

func CompareConfigs(a *UnmarshaledConfig, b *UnmarshaledConfig, scrubSecrets bool) bool {
	aBytes, err := MarshalConfig(a, scrubSecrets)
	if err != nil {
		return false
	}
	bBytes, err := MarshalConfig(b, scrubSecrets)
	if err != nil {
		return false
	}
	return bytes.Equal(aBytes, bBytes)
}

func MarshalConfigToWriter(c *UnmarshaledConfig, w io.Writer, scrubSecrets bool) error {
	enc := yaml.NewEncoder(w)
	defer enc.Close()

	if !scrubSecrets {
		enc.SetHook(func(in interface{}) (ok bool, out interface{}, err error) {
			switch v := in.(type) {
			case config_util.Secret:
				return true, string(v), nil
			case *config_util.URL:
				return true, v.String(), nil
			default:
				return false, nil, nil
			}
		})
	}

	return enc.Encode(c)
}

func MarshalConfig(c *UnmarshaledConfig, scrubSecrets bool) ([]byte, error) {
	var buf bytes.Buffer
	err := MarshalConfigToWriter(c, &buf, scrubSecrets)
	return buf.Bytes(), err
}
