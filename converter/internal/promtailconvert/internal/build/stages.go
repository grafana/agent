package build

import (
	"fmt"

	"github.com/grafana/agent/component/loki/process/stages"
	"github.com/grafana/agent/converter/diag"
	promtailstages "github.com/grafana/loki/clients/pkg/logentry/stages"
	"github.com/mitchellh/mapstructure"
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
		//TODO(thampiotr): add support for all the other stages
		if name == promtailstages.StageTypeJSON {
			return convertJSONStage(iCfg, diags)
		}
	}

	diags.Add(diag.SeverityLevelError, fmt.Sprintf("unsupported pipeline stage: %v", st))
	return stages.StageConfig{}, false
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
