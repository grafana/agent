package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

const (
	StageTypeDiff     = "diff"
	StageTypePipeline = "pipeline"
)

// Processor takes an existing set of labels, timestamp and log entry and returns either a possibly mutated
// timestamp and log entry
type Processor interface {
	Process(labels model.LabelSet, extracted map[string]interface{}, time *time.Time, entry *string)
	Name() string
}

// Stage can receive entries via an inbound channel and forward mutated entries to an outbound channel.
type Stage interface {
	Name() string
	Run(chan Entry) chan Entry
}

func (entry *Entry) copy() *Entry {
	out, err := yaml.Marshal(entry)
	if err != nil {
		return nil
	}

	var n *Entry
	err = yaml.Unmarshal(out, &n)
	if err != nil {
		return nil
	}

	return n
}

// New creates a new stage for the given type and configuration.
func New(logger log.Logger, jobName *string, cfg StageConfig, registerer prometheus.Registerer) (Stage, error) {
	var (
		s   Stage
		err error
	)
	switch {
	case cfg.DiffConfig != nil:
		s, err = newDiffStage(logger, *cfg.DiffConfig)
		if err != nil {
			return nil, err
		}
	default:
		panic("unreacheable; should have decoded into one of the StageConfig fields")
	}
	return s, nil
}
