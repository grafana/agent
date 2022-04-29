package usagestats

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	prom "github.com/prometheus/prometheus/web/api/v1"
)

// ClusterSeed identified a unique cluster
type ClusterSeed struct {
	UID                    string    `json:"UID"`
	CreatedAt              time.Time `json:"created_at"`
	prom.PrometheusVersion `json:"version"`
}

// JSONCodec works as an interface to encode/decode JSON
var JSONCodec = jsonCodec{}

type jsonCodec struct{}

// Decode decodes a JSON
func (jsonCodec) Decode(data []byte) (interface{}, error) {
	var seed ClusterSeed
	if err := jsoniter.ConfigFastest.Unmarshal(data, &seed); err != nil {
		return nil, err
	}
	return &seed, nil
}

// Encode encodes a JSON
func (jsonCodec) Encode(obj interface{}) ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(obj)
}

// CodecID is the ID for the usage stats encode/decoder
func (jsonCodec) CodecID() string { return "usagestats.jsonCodec" }
