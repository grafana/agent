package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
)

const (
	// ErrEmptyLabelDropStageConfig error returned if config is empty
	ErrEmptyLabelDropStageConfig = "labeldrop stage config cannot be empty"
)

// LabelDropConfig is a slice of labels to be dropped
type LabelDropConfig struct {
	Values []string `river:"values,attr"`
}

func validateLabelDropConfig(c LabelDropConfig) error {
	if c.Values == nil || len(c.Values) < 1 {
		return errors.New(ErrEmptyLabelDropStageConfig)
	}

	return nil
}

func newLabelDropStage(config *LabelDropConfig) (Stage, error) {
	err := validateLabelDropConfig(*config)
	if err != nil {
		return nil, err
	}

	return toStage(&labelDropStage{
		config: *config,
	}), nil
}

type labelDropStage struct {
	config LabelDropConfig
}

// Process implements Stage
func (l *labelDropStage) Process(labels model.LabelSet, extracted map[string]interface{}, t *time.Time, entry *string) {
	for _, label := range l.config.Values {
		delete(labels, model.LabelName(label))
	}
}

// Name implements Stage
func (l *labelDropStage) Name() string {
	return StageTypeLabelDrop
}
