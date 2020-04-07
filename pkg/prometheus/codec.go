package prometheus

import (
	"encoding/json"
	"fmt"

	"github.com/cortexproject/cortex/pkg/ring/kv/codec"
)

// GetCodec returns the codec for encoding and decoding InstanceConfigs
func GetCodec() codec.Codec {
	return &jsonCodec{}
}

// TODO(rfratto): add in compression here
type jsonCodec struct{}

func (*jsonCodec) Decode(bb []byte) (interface{}, error) {
	var inst InstanceConfig
	if err := json.Unmarshal(bb, &inst); err != nil {
		return nil, err
	}
	return inst, nil
}

func (*jsonCodec) Encode(v interface{}) ([]byte, error) {
	switch v.(type) {
	case InstanceConfig, *InstanceConfig:
		break
	default:
		panic(fmt.Sprintf("unexpected type %T passed to jsonCodec.Encode", v))
	}

	return json.Marshal(v)
}

func (*jsonCodec) CodecID() string {
	return "agentConfig/json"
}
