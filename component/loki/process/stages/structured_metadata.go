package stages

import (
	"github.com/go-kit/log"
	"github.com/prometheus/common/model"

	"github.com/grafana/loki/pkg/logproto"
)

func newStructuredMetadataStage(logger log.Logger, configs LabelsConfig) (Stage, error) {
	err := validateLabelsConfig(configs)
	if err != nil {
		return nil, err
	}
	return &structuredMetadataStage{
		cfgs:   configs,
		logger: logger,
	}, nil
}

type structuredMetadataStage struct {
	cfgs   LabelsConfig
	logger log.Logger
}

func (s *structuredMetadataStage) Name() string {
	return StageTypeStructuredMetadata
}

func (s *structuredMetadataStage) Run(in chan Entry) chan Entry {
	return RunWith(in, func(e Entry) Entry {
		processLabelsConfigs(s.logger, e.Extracted, s.cfgs, func(labelName model.LabelName, labelValue model.LabelValue) {
			e.StructuredMetadata = append(e.StructuredMetadata, logproto.LabelAdapter{Name: string(labelName), Value: string(labelValue)})
		})
		return e
	})
}
