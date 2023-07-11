package wal

import (
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
)

type MetricStore[T StorableSignal] interface {
	Write(value T) (uint64, error)
	Get(key uint64, value T) bool
	AllKeys() []uint64
	NextKeyValue(start uint64) (uint64, T)

	RegisterTTLCallback(f func(oldestID uint64))
}

type BookmarkStore interface {
	WriteBookmark(wal string, forwardTo string, value *Bookmark) error
	GetBookmark(wal string, forwardTo string) (*Bookmark, bool)
}

type Bookmark struct {
	CurrentIndex uint64
}

type LabelCacheStore interface {
	WriteSignalCache(key string, value *LabelCache) error
	GetSignalCache(key string) (*LabelCache, bool)
}

type LabelCache struct {
	LabelStoreID uint64
	NewestID     uint64
}

type LabelStore interface {
	Write(lbls labels.Labels) uint64
	Get(key uint64) (labels.Labels, bool)
	Delete(key uint64)
}

type StorableSignal interface {
	*Sample | *Exemplar | *Histogram | *FloatHistogram | *Metadata
}

type Sample struct {
	LabelID   uint64
	Timestamp int64
	Value     float64
}

type Exemplar struct {
	Sample
	ExemplarLabelID uint64
}

type Histogram struct {
	LabelID   uint64
	Timestamp int64
	Value     *histogram.Histogram
}

type FloatHistogram struct {
	LabelID   uint64
	Timestamp int64
	Value     *histogram.FloatHistogram
}

type Metadata struct {
	Name    string
	LabelID uint64
	Meta    metadata.Metadata
}
