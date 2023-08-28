package stages

import (
	"github.com/go-kit/log"
	"github.com/prometheus/common/model"

	"github.com/grafana/loki/pkg/logproto"
)

func newNonIndexedLabelsStage(logger log.Logger, configs LabelsConfig) (Stage, error) {
	err := validateLabelsConfig(configs)
	if err != nil {
		return nil, err
	}
	return &nonIndexedLabelsStage{
		cfgs:   configs,
		logger: logger,
	}, nil
}

type nonIndexedLabelsStage struct {
	cfgs   LabelsConfig
	logger log.Logger
}

func (s *nonIndexedLabelsStage) Name() string {
	return StageTypeNonIndexedLabels
}

func (s *nonIndexedLabelsStage) Run(in chan Entry) chan Entry {
	return RunWith(in, func(e Entry) Entry {
		processLabelsConfigs(s.logger, e.Extracted, s.cfgs, func(labelName model.LabelName, labelValue model.LabelValue) {
			e.NonIndexedLabels = append(e.NonIndexedLabels, logproto.LabelAdapter{Name: string(labelName), Value: string(labelValue)})
		})
		return e
	})
}
