package stages

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	json "github.com/json-iterator/go"
	"github.com/prometheus/common/model"
)

// ErrEmptyValuesJSONMergeStageConfig error returned if the config is empty.
var ErrEmptyValuesJSONMergeStageConfig = errors.New("values in jsonmerge config cannot be empty")
var ErrEmptyOutputJSONMergeStageConfig = errors.New("output in jsonmerge config cannot be empty")

// JSONMergeConfig contains the slice of labels to be dropped.
type JSONMergeConfig struct {
	Source string   `river:"source,attr"`
	Values []string `river:"values,attr"`
	Output string   `river:"output,attr"`
}

func newJSONMergeStage(logger log.Logger, config JSONMergeConfig) (Stage, error) {
	if len(config.Values) < 1 {
		return nil, ErrEmptyValuesJSONMergeStageConfig
	}

	if config.Output == "" {
		return nil, ErrEmptyOutputJSONMergeStageConfig
	}

	return toStage(&jsonmergeStage{
		config: config,
		logger: logger,
	}), nil
}

type jsonmergeStage struct {
	config JSONMergeConfig
	logger log.Logger
}

// Process implements Stage.
func (j *jsonmergeStage) Process(labels model.LabelSet, extracted map[string]interface{}, t *time.Time, entry *string) {
	obj, err := j.extractSourceObject(extracted, entry)
	if err != nil {
		if Debug {
			level.Debug(j.logger).Log("msg", "unable to extract source value", "error", err.Error())
		}
	}

	for _, value := range j.config.Values {
		extractedValue, ok := extracted[value]
		if !ok {
			if Debug {
				level.Debug(j.logger).Log("msg", "unable to find value from extracted map", "value", value)
			}
			continue
		}
		obj[value] = extractedValue
	}

	b, err := json.Marshal(obj)
	if err != nil {
		if Debug {
			level.Debug(j.logger).Log("msg", "unable to json marshal merged object", "error", err.Error())
		}
		return
	}

	extracted[j.config.Output] = string(b)
}

// Name implements Stage.
func (j *jsonmergeStage) Name() string {
	return StageTypeLabelDrop
}

func (j *jsonmergeStage) extractSourceObject(extracted map[string]interface{}, entry *string) (map[string]interface{}, error) {
	if j.config.Source == "" {
		b := []byte(*entry)
		var obj map[string]interface{}
		if err := json.Unmarshal(b, &obj); err != nil {
			return nil, fmt.Errorf("unable to json unmarshal extracted source %s: %w", j.config.Source, err)
		}
		return obj, nil
	}

	v, ok := extracted[j.config.Source]
	if !ok {
		return nil, fmt.Errorf("unable to find %s in extracted map", j.config.Source)
	}
	return map[string]interface{}{j.config.Source: v}, nil
}
