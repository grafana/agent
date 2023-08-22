package simple

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/grafana/agent/component/prometheus"
	"time"
)

type SignalDB interface {
	GetNewKey() uint64
	GetOldestKey() uint64
	GetKeys() ([]uint64, error)
	GetCurrentKey() uint64
	GetNextKey(k uint64) uint64
	DeleteKeysOlderThan(k uint64)

	GetValueByByte(k []byte) (any, bool, error)
	GetValueByString(k string) (any, bool, error)
	GetValueByUint(k uint64) (any, bool, error)

	WriteValueWithAutokey(data any, ttl time.Duration) (uint64, error)
	WriteValue(key []byte, data any, ttl time.Duration) error

	Evict() error
	Size() uint64
}

func GetType(data any) (int8, error) {
	switch v := data.(type) {
	case []prometheus.Sample:
		return MetricSignal, nil
	case []prometheus.Exemplar:
		return ExemplarSignal, nil
	case []prometheus.Metadata:
		return MetadataSignal, nil
	case []prometheus.Histogram:
		return HistogramSignal, nil
	case []prometheus.FloatHistogram:
		return FloathistogramSignal, nil
	case *Bookmark:
		return BookmarkType, nil
	default:
		return 0, fmt.Errorf("unknown data type %v", v)
	}
}

func GetValue(data []byte, t int8) any {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	switch t {
	case MetricSignal:
		var val []prometheus.Sample
		dec.Decode(&val)
		return val
	case ExemplarSignal:
		var val []prometheus.Exemplar
		dec.Decode(&val)
		return val
	case MetadataSignal:
		var val []prometheus.Metadata
		dec.Decode(&val)
		return val
	case HistogramSignal:
		var val []prometheus.Histogram
		dec.Decode(&val)
		return val
	case FloathistogramSignal:
		var val []prometheus.FloatHistogram
		dec.Decode(&val)
		return val
	case BookmarkType:
		val := &Bookmark{}
		dec.Decode(val)
		return val
	default:
		return nil
	}
}
