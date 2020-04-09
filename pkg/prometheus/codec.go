package prometheus

import (
	"bytes"
	"compress/gzip"
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
	r, err := gzip.NewReader(bytes.NewReader(bb))
	if err != nil {
		return nil, err
	}

	var inst InstanceConfig
	if err := json.NewDecoder(r).Decode(&inst); err != nil {
		return nil, err
	}
	return inst, nil
}

func (*jsonCodec) Encode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer

	switch v.(type) {
	case InstanceConfig, *InstanceConfig:
		break
	default:
		panic(fmt.Sprintf("unexpected type %T passed to jsonCodec.Encode", v))
	}

	w := gzip.NewWriter(&buf)
	jsonEncoder := json.NewEncoder(w)
	if err := jsonEncoder.Encode(v); err != nil {
		return nil, err
	}

	w.Close()
	return buf.Bytes(), nil
}

func (*jsonCodec) CodecID() string {
	return "agentConfig/json"
}
