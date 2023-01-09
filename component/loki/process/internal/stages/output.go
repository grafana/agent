package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"errors"
	"reflect"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
)

// Config Errors
const (
	ErrEmptyOutputStageConfig = "output stage config cannot be empty"
	ErrOutputSourceRequired   = "output source value is required if output is specified"
)

// OutputConfig represents an Output Stage configuration which sets the log
// line to an entry of the the extracted value map.
type OutputConfig struct {
	Source string `river:"source,attr"`
}

// validateOutput validates the outputStage config
func validateOutputConfig(cfg *OutputConfig) error {
	if cfg == nil {
		return errors.New(ErrEmptyOutputStageConfig)
	}
	if cfg.Source == "" {
		return errors.New(ErrOutputSourceRequired)
	}
	return nil
}

// newOutputStage creates a new outputStage
func newOutputStage(logger log.Logger, config *OutputConfig) (Stage, error) {
	err := validateOutputConfig(config)
	if err != nil {
		return nil, err
	}
	return toStage(&outputStage{
		config: config,
		logger: logger,
	}), nil
}

// outputStage will mutate the incoming entry and set it from extracted data
type outputStage struct {
	config *OutputConfig
	logger log.Logger
}

// Process implements Stage
func (o *outputStage) Process(labels model.LabelSet, extracted map[string]interface{}, t *time.Time, entry *string) {
	if o.config == nil {
		return
	}
	if v, ok := extracted[o.config.Source]; ok {
		s, err := getString(v)
		if err != nil {
			level.Debug(o.logger).Log("msg", "extracted output could not be converted to a string", "err", err, "type", reflect.TypeOf(v))
			return
		}
		*entry = s
	} else {
		level.Debug(o.logger).Log("msg", "extracted data did not contain output source")
	}
}

// Name implements Stage
func (o *outputStage) Name() string {
	return StageTypeOutput
}
