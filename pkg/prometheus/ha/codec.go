package ha

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/cortexproject/cortex/pkg/ring/kv/codec"
	"github.com/grafana/agent/pkg/prometheus/instance"
	"gopkg.in/yaml.v2"
)

// GetCodec returns the codec for encoding and decoding instance.Configs
func GetCodec() codec.Codec {
	return &yamlCodec{}
}

type yamlCodec struct{}

func (*yamlCodec) Decode(bb []byte) (interface{}, error) {
	r, err := gzip.NewReader(bytes.NewReader(bb))
	if err != nil {
		return nil, err
	}

	var inst instance.Config
	if err := yaml.NewDecoder(r).Decode(&inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

func (*yamlCodec) Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer

	switch v.(type) {
	case instance.Config, *instance.Config:
		break
	default:
		panic(fmt.Sprintf("unexpected type %T passed to yamlCodec.Encode", v))
	}

	w := gzip.NewWriter(&buf)
	yamlEncoder := yaml.NewEncoder(w)
	if err := yamlEncoder.Encode(v); err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

func (*yamlCodec) CodecID() string {
	return "agentConfig/yaml"
}
