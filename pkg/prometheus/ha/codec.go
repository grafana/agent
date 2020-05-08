package ha

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/cortexproject/cortex/pkg/ring/kv/codec"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"github.com/prometheus/common/config"
	"gopkg.in/yaml.v2"
)

// GetCodec returns the codec for encoding and decoding instance.Configs
func GetCodec() codec.Codec {
	return &yamlCodec{}
}

type yamlCodec struct{}

func (*yamlCodec) Decode(bb []byte) (interface{}, error) {
	// Decode is called by kv.Clients with an empty slice when a
	// key is deleted. We should stop early here and don't return
	// an error so the deletion event propagates to watchers.
	if len(bb) == 0 {
		return nil, nil
	}

	r, err := gzip.NewReader(bytes.NewReader(bb))
	if err != nil {
		return nil, err
	}

	var codecConfig codecConfig
	if err := yaml.NewDecoder(r).Decode(&codecConfig); err != nil {
		return nil, err
	}
	restoreConfig(&codecConfig)
	return codecConfig.Config, nil
}

func (*yamlCodec) Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer

	var configToWrite *instance.Config

	switch val := v.(type) {
	case instance.Config:
		configToWrite = &val
	case *instance.Config:
		configToWrite = val
	default:
		panic(fmt.Sprintf("unexpected type %T passed to yamlCodec.Encode", v))
	}

	codecConfig := newCodecConfig(configToWrite)

	w := gzip.NewWriter(&buf)
	yamlEncoder := yaml.NewEncoder(w)
	if err := yamlEncoder.Encode(codecConfig); err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

func (*yamlCodec) CodecID() string {
	return "agentConfig/yaml"
}

type codecConfig struct {
	Config  *instance.Config `yaml:"config"`
	Secrets []configSecrets  `yaml:"secrets"`
}

// newCodecConfig creates a new codecConfig that is ready for marshaling as YAML
// and storing in a KV store. newCodecConfig must extract secrets stored in the
// instance.Config and store them elsewhere as strings as the MarshalYAML function
// on the Prometheus Secret type replaces the contents of the secret with
// the text "<secret>".
func newCodecConfig(c *instance.Config) *codecConfig {
	var secrets []configSecrets
	for _, rwr := range c.RemoteWrite {
		var s configSecrets
		if rwr.HTTPClientConfig.BasicAuth != nil {
			s.BasicAuthPassword = string(rwr.HTTPClientConfig.BasicAuth.Password)
		}
		s.BearerToken = string(rwr.HTTPClientConfig.BearerToken)
		secrets = append(secrets, s)
	}
	return &codecConfig{
		Config:  c,
		Secrets: secrets,
	}
}

type configSecrets struct {
	BasicAuthPassword string `yaml:"basic_auth_password"`
	BearerToken       string `yaml:"bearer_token"`
}

// restoreConfig restores the instance config stored in a codec config by copying
// extracted secrets to it.
func restoreConfig(c *codecConfig) {
	for i, rwr := range c.Config.RemoteWrite {
		if rwr.HTTPClientConfig.BasicAuth != nil {
			rwr.HTTPClientConfig.BasicAuth.Password = config.Secret(c.Secrets[i].BasicAuthPassword)
		}
		rwr.HTTPClientConfig.BearerToken = config.Secret(c.Secrets[i].BearerToken)
	}
}
