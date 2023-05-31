package prometheus

import (
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
)

// WriteTo is an interface used by the Watcher to send the samples it's read
// from the WAL on to somewhere else. Functions will be called concurrently
// and it is left to the implementer to make sure they are safe.
type WriteTo interface {
	Append([]Sample) bool
	AppendExemplars([]Exemplar) bool
	AppendHistograms([]Histogram) bool
	AppendFloatHistograms([]FloatHistogram) bool
	AppendMetadata([]Metadata) bool
}

type Sample struct {
	GlobalRefID uint64
	Timestamp   int64
	Value       float64
}

type Exemplar struct {
	Sample
	L labels.Labels
}

type Histogram struct {
	GlobalRefID uint64
	Timestamp   int64
	Value       *histogram.Histogram
}

type FloatHistogram struct {
	GlobalRefID uint64
	Timestamp   int64
	Value       *histogram.FloatHistogram
}

type Metadata struct {
	Name        string
	GlobalRefID uint64
	Meta        metadata.Metadata
}
