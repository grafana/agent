package config

import (
	"time"

	"github.com/alecthomas/units"
	"github.com/prometheus/common/model"
	internal "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/labels"
)

type GlobalConfig struct {
	scrapeInterval        time.Duration     `river:"scrape_interval,attr,optional"`
	scrapeTimeout         time.Duration     `river:"scrape_timeout,attr,optional"`
	evaluationInterval    time.Duration     `river:"evaluation_interval,attr,optional"`
	queryLogFile          string            `river:"query_log_file,string,optional"`
	externalLabels        map[string]string `river:"external_labels,attr,optional"`
	bodySizeLimit         string            `river:"body_size_limit,string,optional"`
	sampleLimit           uint              `river:"sample_limit,number,optional"`
	targetLimit           uint              `river:"target_limit,number,optional"`
	labelLimit            uint              `river:"label_limit,number,optional"`
	labelNameLengthLimit  uint              `river:"label_name_length_limit,number,optional"`
	labelValueLengthLimit uint              `river:"label_value_length_limit,number,optional"`
}

func (config *GlobalConfig) ToInternal() (*internal.GlobalConfig, error) {
	body_size_limit, err := units.ParseBase2Bytes(config.bodySizeLimit)
	if err != nil {
		return nil, err
	}

	return &internal.GlobalConfig{
		ScrapeInterval:        model.Duration(config.scrapeInterval),
		ScrapeTimeout:         model.Duration(config.scrapeTimeout),
		EvaluationInterval:    model.Duration(config.evaluationInterval),
		QueryLogFile:          config.queryLogFile,
		ExternalLabels:        config.externalLabelsToInternals(),
		BodySizeLimit:         body_size_limit,
		SampleLimit:           config.sampleLimit,
		TargetLimit:           config.targetLimit,
		LabelLimit:            config.labelLimit,
		LabelNameLengthLimit:  config.labelNameLengthLimit,
		LabelValueLengthLimit: config.labelValueLengthLimit,
	}, nil
}

func (config *GlobalConfig) externalLabelsToInternals() labels.Labels {
	var ls labels.Labels

	for name, value := range config.externalLabels {
		ls = append(ls, labels.Label{
			Name:  name,
			Value: value,
		})
	}
	return ls
}
