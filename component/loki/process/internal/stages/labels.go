package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"fmt"
	"reflect"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
)

const (
	ErrEmptyLabelStageConfig = "label stage config cannot be empty"
	ErrInvalidLabelName      = "invalid label name: %s"
)

// LabelsConfig is a set of labels to be extracted
type LabelsConfig struct {
	Values map[string]*string `river:"values,attr"`
}

// validateLabelsConfig validates the Label stage configuration
func validateLabelsConfig(c LabelsConfig) error {
	if c.Values == nil {
		return errors.New(ErrEmptyLabelStageConfig)
	}
	for labelName, labelSrc := range c.Values {
		if !model.LabelName(labelName).IsValid() {
			return fmt.Errorf(ErrInvalidLabelName, labelName)
		}
		// If no label source was specified, use the key name
		if labelSrc == nil || *labelSrc == "" {
			lName := labelName
			c.Values[labelName] = &lName
		}
	}
	return nil
}

// newLabelStage creates a new label stage to set labels from extracted data
func newLabelStage(logger log.Logger, configs LabelsConfig) (Stage, error) {
	err := validateLabelsConfig(configs)
	if err != nil {
		return nil, err
	}
	return toStage(&labelStage{
		cfgs:   configs,
		logger: logger,
	}), nil
}

// labelStage sets labels from extracted data
type labelStage struct {
	cfgs   LabelsConfig
	logger log.Logger
}

// Process implements Stage
func (l *labelStage) Process(labels model.LabelSet, extracted map[string]interface{}, t *time.Time, entry *string) {
	for lName, lSrc := range l.cfgs.Values {
		if lValue, ok := extracted[*lSrc]; ok {
			s, err := getString(lValue)
			if err != nil {
				if Debug {
					level.Debug(l.logger).Log("msg", "failed to convert extracted label value to string", "err", err, "type", reflect.TypeOf(lValue))
				}
				continue
			}
			labelValue := model.LabelValue(s)
			if !labelValue.IsValid() {
				if Debug {
					level.Debug(l.logger).Log("msg", "invalid label value parsed", "value", labelValue)
				}
				continue
			}
			labels[model.LabelName(lName)] = labelValue
		}
	}
}

// Name implements Stage
func (l *labelStage) Name() string {
	return StageTypeLabel
}
