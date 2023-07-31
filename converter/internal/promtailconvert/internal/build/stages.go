package build

import (
	"fmt"
	"sort"
	"time"

	"github.com/alecthomas/units"

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
		case promtailstages.StageTypeLabel:
			return convertLabels(iCfg, diags)
		case promtailstages.StageTypeLabelDrop:
			return convertLabelDrop(iCfg, diags)
		case promtailstages.StageTypeTimestamp:
			return convertTimestamp(iCfg, diags)
		case promtailstages.StageTypeOutput:
			return convertOutput(iCfg, diags)
		case promtailstages.StageTypeDocker:
			return convertDocker()
		case promtailstages.StageTypeCRI:
			return convertCRI()
		case promtailstages.StageTypeMatch:
			return convertMatch(iCfg, diags)
		case promtailstages.StageTypeTemplate:
			return convertTemplate(iCfg, diags)
		case promtailstages.StageTypeTenant:
			return convertTenant(iCfg, diags)
		case promtailstages.StageTypeDrop:
			return convertDrop(iCfg, diags)
		case promtailstages.StageTypeSampling:
			return convertSampling(iCfg, diags)
		case promtailstages.StageTypeLimit:
			return convertLimit(iCfg, diags)
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

func convertLimit(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pLimit := &promtailstages.LimitConfig{}
	if err := mapstructure.Decode(cfg, pLimit); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{
		LimitConfig: &stages.LimitConfig{
			Rate:              pLimit.Rate,
			Burst:             pLimit.Burst,
			Drop:              pLimit.Drop,
			ByLabelName:       pLimit.ByLabelName,
			MaxDistinctLabels: pLimit.MaxDistinctLabels,
		},
	}, true
}

func convertSampling(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	diags.Add(diag.SeverityLevelError, fmt.Sprintf("pipeline_stages.sampling is currently not supported: %v", cfg))
	return stages.StageConfig{}, false
}

func convertDrop(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pDrop := &promtailstages.DropConfig{}
	if err := mapstructure.Decode(cfg, pDrop); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}

	source := ""
	if pDrop.Source != nil {
		switch s := pDrop.Source.(type) {
		case []interface{}:
			if len(s) == 1 {
				str, ok := s[0].(string)
				if !ok {
					diags.Add(
						diag.SeverityLevelError,
						fmt.Sprintf("invalid pipeline_stages.drop.source[0] field type '%T': %v", s[0], s[0]),
					)
					return stages.StageConfig{}, false
				}
				source = str
			} else if len(s) > 1 {
				diags.Add(
					diag.SeverityLevelError,
					fmt.Sprintf("only single value for pipelina_stages.drop.source is supported - got: %v", s),
				)
				return stages.StageConfig{}, false
			}
		case string:
			source = s
		default:
			diags.Add(
				diag.SeverityLevelError,
				fmt.Sprintf("invalid pipeline_stages.drop.source field type '%T': %v", pDrop.Source, pDrop.Source),
			)
			return stages.StageConfig{}, false
		}
	}

	var olderThan time.Duration
	if pDrop.OlderThan != nil {
		d, err := time.ParseDuration(*pDrop.OlderThan)
		if err != nil {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("invalid pipeline_stages.drop.older_than field: %v", err))
			return stages.StageConfig{}, false
		}
		olderThan = d
	}

	var longerThan units.Base2Bytes
	if pDrop.LongerThan != nil {
		lt, err := units.ParseBase2Bytes(*pDrop.LongerThan)
		if err != nil {
			diags.Add(diag.SeverityLevelError, fmt.Sprintf("invalid pipeline_stages.drop.longer_than field: %v", err))
			return stages.StageConfig{}, false
		}
		longerThan = lt
	}

	if pDrop.Separator != nil && *pDrop.Separator != "" {
		diags.Add(
			diag.SeverityLevelWarn,
			fmt.Sprintf("pipeline_stages.drop.separator is ignored since only one 'source' value is supported: %v", *pDrop.Separator),
		)
	}

	return stages.StageConfig{DropConfig: &stages.DropConfig{
		DropReason: defaultEmpty(pDrop.DropReason),
		Source:     source,
		Value:      defaultEmpty(pDrop.Value),
		Expression: defaultEmpty(pDrop.Expression),
		OlderThan:  olderThan,
		LongerThan: longerThan,
	}}, true
}

func convertTenant(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pTenant := &promtailstages.TenantConfig{}
	if err := mapstructure.Decode(cfg, pTenant); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{TenantConfig: &stages.TenantConfig{
		Label:  pTenant.Label,
		Source: pTenant.Source,
		Value:  pTenant.Value,
	}}, true
}

func convertTemplate(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pTemplate := &promtailstages.TemplateConfig{}
	if err := mapstructure.Decode(cfg, pTemplate); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{TemplateConfig: &stages.TemplateConfig{
		Source:   pTemplate.Source,
		Template: pTemplate.Template,
	}}, true
}

func convertMatch(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pMatch := &promtailstages.MatcherConfig{}
	if err := mapstructure.Decode(cfg, pMatch); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}

	// convert nested stages
	subStages := make([]stages.StageConfig, len(pMatch.Stages))
	for i, ps := range pMatch.Stages {
		if fs, ok := convertStage(ps, diags); ok {
			subStages[i] = fs
		}
	}

	return stages.StageConfig{MatchConfig: &stages.MatchConfig{
		Selector:     pMatch.Selector,
		Stages:       subStages,
		Action:       pMatch.Action,
		PipelineName: defaultEmpty(pMatch.PipelineName),
		DropReason:   defaultEmpty(pMatch.DropReason),
	}}, true
}

func convertCRI() (stages.StageConfig, bool) {
	return stages.StageConfig{CRIConfig: &stages.CRIConfig{}}, true
}

func convertDocker() (stages.StageConfig, bool) {
	return stages.StageConfig{DockerConfig: &stages.DockerConfig{}}, true
}

func convertOutput(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pOutput := &promtailstages.OutputConfig{}
	if err := mapstructure.Decode(cfg, pOutput); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{OutputConfig: &stages.OutputConfig{
		Source: pOutput.Source,
	}}, true
}

func convertTimestamp(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pTimestamp := &promtailstages.TimestampConfig{}
	if err := mapstructure.Decode(cfg, pTimestamp); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{TimestampConfig: &stages.TimestampConfig{
		Source:          pTimestamp.Source,
		Format:          pTimestamp.Format,
		FallbackFormats: pTimestamp.FallbackFormats,
		Location:        pTimestamp.Location,
		ActionOnFailure: defaultEmpty(pTimestamp.ActionOnFailure),
	},
	}, true
}

func convertLabelDrop(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pLabelDrop := &promtailstages.LabelDropConfig{}
	if err := mapstructure.Decode(cfg, pLabelDrop); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{LabelDropConfig: &stages.LabelDropConfig{
		Values: *pLabelDrop,
	}}, true
}

func convertLabels(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pLabels := &promtailstages.LabelsConfig{}
	if err := mapstructure.Decode(cfg, pLabels); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}
	return stages.StageConfig{LabelsConfig: &stages.LabelsConfig{
		Values: *pLabels,
	}}, true
}

func convertMetrics(cfg interface{}, diags *diag.Diagnostics) (stages.StageConfig, bool) {
	pMetrics := &promtailstages.MetricsConfig{}
	if err := mapstructure.Decode(cfg, pMetrics); err != nil {
		addInvalidStageError(diags, cfg, err)
		return stages.StageConfig{}, false
	}

	var fMetrics []stages.MetricConfig

	// sort metric names to make conversion deterministic
	var sortedNames []string
	for name := range *pMetrics {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		pMetric := (*pMetrics)[name]
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

// unifySourceField reflects the implementation from promtail's clients/pkg/logentry/stages/drop.go
func unifySourceField(i interface{}, diags *diag.Diagnostics) string {
	if i == nil {
		return ""
	}

	switch s := i.(type) {
	case []string:
		if len(s) == 0 {
			return ""
		}
		if len(s) == 1 {
			return s[0]
		}
		diags.Add(diag.SeverityLevelError, fmt.Sprintf("only single value for pipelina_stages.drop.source is supported - got: %v", s))
		return ""
	case string:
		return s
	}

	diags.Add(diag.SeverityLevelError, fmt.Sprintf("invalid source field type: %T - %v", i, i))
	return ""
}
