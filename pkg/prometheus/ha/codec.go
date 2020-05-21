package ha

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/cortexproject/cortex/pkg/ring/kv/codec"
	"github.com/grafana/agent/pkg/prometheus/instance"
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

	return instance.UnmarshalConfig(r)
}

func (*yamlCodec) Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer

	var cfg *instance.Config

	switch v := v.(type) {
	case instance.Config:
		cfg = &v
	case *instance.Config:
		cfg = v
	default:
		panic(fmt.Sprintf("unexpected type %T passed to yamlCodec.Encode", v))
	}

	w := gzip.NewWriter(&buf)
	err := instance.MarshalConfigToWriter(cfg, w, false)
	if err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

func (*yamlCodec) CodecID() string {
	return "agentConfig/yaml"
}
