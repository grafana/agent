package build

import (
	"fmt"
	"time"

	promtailmetric "github.com/grafana/loki/clients/pkg/logentry/metric"
	promtailstages "github.com/grafana/loki/clients/pkg/logentry/stages"
	"github.com/mitchellh/mapstructure"

	"github.com/grafana/agent/component/loki/process/metric"
	"github.com/grafana/agent/component/loki/process/stages"
	"github.com/grafana/agent/converter/diag"
)

func convertStage(st interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	stage, ok := st.(promtailstages.PipelineStage)
	if !ok {
		diags.Add(diag.SeverityLevelCritical, fmt.Sprintf("invalid input YAML config, "+
			"make sure each stage of your pipeline is a YAML object (must end with a `:`), check stage `- %s`", st))
		return stages.StageConfig{}, false
	}
	if len(stage) != 1 {
		diags.Add(
			diag.SeverityLevelCritical,
			fmt.Sprintf("each pipeline stage must contain exactly one key, got: %v", stage),
		)
		return stages.StageConfig{}, false
	}

	for iName, iCfg := range stage {
		name, ok := iName.(string)
		if !ok {
			addInvalidStageError(diags, iCfg, fmt.Errorf("stage name must be a string, got %T", iName))
		}

		switch name {
		case promtailstages.StageTypeJSON:
			return convertJSONStage(iCfg, diags)
		case promtailstages.StageTypeLogfmt:
			return convertLogfmt(iCfg, diags)
		case promtailstages.StageTypeRegex:
			return convertRegex(iCfg, diags)
		case promtailstages.StageTypeReplace:
			return convertReplace(iCfg, diags)
		case promtailstages.StageTypeMetric:
			return convertMetrics(iCfg, diags)
			//case promtailstages.StageTypeLabel:
			//	return convertlabels(iCfg, diags)
			//case promtailstages.StageTypeLabelDrop:
			//	return convertlabeldrop(iCfg, diags)
			//case promtailstages.StageTypeTimestamp:
			//	return converttimestamp(iCfg, diags)
			//case promtailstages.StageTypeOutput:
			//	return convertoutput(iCfg, diags)
			//case promtailstages.StageTypeDocker:
			//	return convertdocker(iCfg, diags)
			//case promtailstages.StageTypeCRI:
			//	return convertcri(iCfg, diags)
			//case promtailstages.StageTypeMatch:
			//	return convertmatch(iCfg, diags)
			//case promtailstages.StageTypeTemplate:
			//	return converttemplate(iCfg, diags)
			//case promtailstages.StageTypePipeline:
			//	return convertpipeline(iCfg, diags)
			//case promtailstages.StageTypeTenant:
			//	return converttenant(iCfg, diags)
			//case promtailstages.StageTypeDrop:
			//	return convertdrop(iCfg, diags)
			//case promtailstages.StageTypeSampling:
			//	return convertsampling(iCfg, diags)
			//case promtailstages.StageTypeLimit:
			//	return convertlimit(iCfg, diags)
			//case promtailstages.StageTypeMultiline:
			//	return convertmultiline(iCfg, diags)
			//case promtailstages.StageTypePack:
			//	return convertpack(iCfg, diags)
			//case promtailstages.StageTypeLabelAllow:
			//	return convertlabelallow(iCfg, diags)
			//case promtailstages.StageTypeStaticLabels:
			//	return convertstatic_labels(iCfg, diags)
			//case promtailstages.StageTypeDecolorize:
			//	return convertdecolorize(iCfg, diags)
			//case promtailstages.StageTypeEventLogMessage:
			//	return converteventlogmessage(iCfg, diags)
			//case promtailstages.StageTypeGeoIP:
			//	return convertgeoip(iCfg, diags)
		}
	}

	diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported pipeline stage: %v", st))
	return stages.StageConfig{}, false
}

func convertMetrics(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pMetrics := &promtailstages.MetricsConfig{}
	if err := mapstructure.Decode(cfg, pMetrics); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}

	var fMetrics []stages.MetricConfig
	for name, pMetric := range *pMetrics {
		fMetric, ok := toFlowMetricProcessStage(name, pMetric, diags)
		if !ok {
			return stages.StageConfig{}, false
		}
		fMetrics = append(fMetrics, fMetric)
	}
	return stages.StageConfig{MetricsConfig: &stages.MetricsConfig{
		Metrics: fMetrics,
	}}, true
}

func toFlowMetricProcessStage(name string, pMetric promtailstages.MetricConfig, diags *diag.Diagnostics) (stages.MetricConfig, bool) {
	var fMetric stages.MetricConfig

	var maxIdle time.Duration
	if pMetric.IdleDuration != nil {
		d, err := time.ParseDuration(*pMetric.IdleDuration)
		if err != nil {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to parse duration: %s - %v", *pMetric.IdleDuration, err))
			return stages.MetricConfig{}, false
		}
		maxIdle = d
	}

	// Create metric according to type
	switch pMetric.MetricType {
	case promtailstages.MetricTypeCounter:
		pCounter, err := promtailmetric.NewCounters(name, pMetric.Description, pMetric.Config, int64(maxIdle.Seconds()))
		if err != nil {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to create counter metric process stage: %v", err))
			return stages.MetricConfig{}, false
		}
		fMetric.Counter = &metric.CounterConfig{
			Name:            name,
			Description:     pMetric.Description,
			Source:          defaultEmpty(pMetric.Source),
			Prefix:          pMetric.Prefix,
			MaxIdle:         maxIdle,
			Value:           defaultEmpty(pCounter.Cfg.Value),
			Action:          pCounter.Cfg.Action,
			MatchAll:        defaultFalse(pCounter.Cfg.MatchAll),
			CountEntryBytes: defaultFalse(pCounter.Cfg.CountBytes),
		}
	case promtailstages.MetricTypeGauge:
		pGauge, err := promtailmetric.NewGauges(name, pMetric.Description, pMetric.Config, int64(maxIdle.Seconds()))
		if err != nil {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to create gauge metric process stage: %v", err))
			return stages.MetricConfig{}, false
		}
		fMetric.Gauge = &metric.GaugeConfig{
			Name:        name,
			Description: pMetric.Description,
			Source:      defaultEmpty(pMetric.Source),
			Prefix:      pMetric.Prefix,
			MaxIdle:     maxIdle,
			Value:       defaultEmpty(pGauge.Cfg.Value),
			Action:      pGauge.Cfg.Action,
		}
	case promtailstages.MetricTypeHistogram:
		pHistogram, err := promtailmetric.NewHistograms(name, pMetric.Description, pMetric.Config, int64(maxIdle.Seconds()))
		if err != nil {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("failed to create histogram metric process stage: %v", err))
			return stages.MetricConfig{}, false
		}
		fMetric.Histogram = &metric.HistogramConfig{
			Name:        name,
			Description: pMetric.Description,
			Source:      defaultEmpty(pMetric.Source),
			Prefix:      pMetric.Prefix,
			MaxIdle:     maxIdle,
			Value:       defaultEmpty(pHistogram.Cfg.Value),
			Buckets:     pHistogram.Cfg.Buckets,
		}
	}
	return fMetric, true
}

func convertReplace(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pCfg := &promtailstages.ReplaceConfig{}
	if err := mapstructure.Decode(cfg, pCfg); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{
		ReplaceConfig: &stages.ReplaceConfig{
			Expression: pCfg.Expression,
			Source:     defaultEmpty(pCfg.Source),
			Replace:    pCfg.Replace,
		}}, true
}

func convertRegex(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pCfg := &promtailstages.RegexConfig{}
	if err := mapstructure.Decode(cfg, pCfg); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{
		RegexConfig: &stages.RegexConfig{
			Expression: pCfg.Expression,
			Source:     pCfg.Source,
		}}, true
}

func convertLogfmt(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pCfg := &promtailstages.LogfmtConfig{}
	if err := mapstructure.Decode(cfg, pCfg); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{
		LogfmtConfig: &stages.LogfmtConfig{
			Source:  defaultEmpty(pCfg.Source),
			Mapping: pCfg.Mapping,
		}}, true
}

func convertJSONStage(iCfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pCfg := &promtailstages.JSONConfig{}
	if err := mapstructure.Decode(iCfg, pCfg); err != nil {
		addInvalidStageError(diags, iCfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{
		JSONConfig: &stages.JSONConfig{
			Expressions:   pCfg.Expressions,
			Source:        pCfg.Source,
			DropMalformed: pCfg.DropMalformed,
		}}, true
}

func addInvalidStageError(diags *diag.Diagnostics, iCfg interface{}, err error) {
	diags.Add(
		diag.SeverityLevelError,
		fmt.Sprintf("invalid pipeline stage config: %v - %v", iCfg, err),
	)
}

func defaultEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func defaultFalse(s *bool) bool {
	if s == nil {
		return false
	}
	return *s
}
