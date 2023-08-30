package prometheus

import (
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
)

type Metadata struct {
	GlobalRefID uint64
	Name        string
	L           labels.Labels
	Meta        metadata.Metadata
}
