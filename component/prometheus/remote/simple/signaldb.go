package simple

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/grafana/agent/component/prometheus"
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
	SeriesCount() int64
}

func GetType(data any) (int8, int, error) {
	switch v := data.(type) {
	case []prometheus.Sample:
		return MetricSignal, len(v), nil
	case []prometheus.Exemplar:
		return ExemplarSignal, len(v), nil
	case []prometheus.Metadata:
		return MetadataSignal, len(v), nil
	case []prometheus.Histogram:
		return HistogramSignal, len(v), nil
	case []prometheus.FloatHistogram:
		return FloathistogramSignal, len(v), nil
	case *Bookmark:
		return BookmarkType, 1, nil
	default:
		return 0, 0, fmt.Errorf("unknown data type %v", v)
	}
}

func GetValue(data []byte, t int8) any {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var err error
	defer func() {
		if err != nil {
			println(err)
		}
	}()
	switch t {
	case MetricSignal:
		var val []prometheus.Sample
		err = dec.Decode(&val)
		return val
	case ExemplarSignal:
		var val []prometheus.Exemplar
		err = dec.Decode(&val)
		return val
	case MetadataSignal:
		var val []prometheus.Metadata
		err = dec.Decode(&val)
		return val
	case HistogramSignal:
		var val []prometheus.Histogram
		err = dec.Decode(&val)
		return val
	case FloathistogramSignal:
		var val []prometheus.FloatHistogram
		err = dec.Decode(&val)
		return val
	case BookmarkType:
		val := &Bookmark{}
		err = dec.Decode(val)
		return val
	default:
		return nil
	}
}
