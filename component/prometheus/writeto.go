package prometheus

import (
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
)

type Sample struct {
	GlobalRefID uint64
	L           labels.Labels
	Timestamp   int64
	Value       float64
}

type Exemplar struct {
	GlobalRefID uint64
	Sample
	L labels.Labels
}

type Histogram struct {
	GlobalRefID uint64
	L           labels.Labels
	Timestamp   int64
	Value       *histogram.Histogram
}

type FloatHistogram struct {
	GlobalRefID uint64
	L           labels.Labels
	Timestamp   int64
	Value       *histogram.FloatHistogram
}

type Metadata struct {
	GlobalRefID uint64
	Name        string
	L           labels.Labels
	Meta        metadata.Metadata
}
