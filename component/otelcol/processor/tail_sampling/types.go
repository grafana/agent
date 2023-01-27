package tail_sampling

import (
	"github.com/mitchellh/mapstructure"
	tsp "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
)

type PolicyCfg struct {
	Name                string              `river:"name,attr"`
	Type                string              `river:"type,attr"`
	LatencyCfg          LatencyCfg          `river:"latency,block,optional"`
	NumericAttributeCfg NumericAttributeCfg `river:"numeric_attribute,block,optional"`

	//sharedPolicyCfg `river:"shared_policy,block"`

	// Configs for defining composite policy
	// TODO CompositeCfg tsp.CompositeCfg `mapstructure:"composite"`

	// Configs for defining and policy
	// TODO AndCfg tsp.AndCfg `mapstructure:"and"`
}

func (args Arguments) convertPolicyCfg() []tsp.PolicyCfg {
	otelCfg := []tsp.PolicyCfg{}

	for _, policy := range args.PolicyCfgs {
		var otelPolicy tsp.PolicyCfg

		err := mapstructure.Decode(map[string]interface{}{
			"name":              policy.Name,
			"type":              policy.Type,
			"latency":           policy.convertLatencyCfg(),
			"numeric_attribute": policy.convertNumericAttributeCfg(),
		}, &otelPolicy)

		if err != nil {
			panic(err)
		}

		otelCfg = append(otelCfg, otelPolicy)
	}

	return otelCfg
}

// LatencyCfg holds the configurable settings to create a latency filter sampling policy
// evaluator
type LatencyCfg struct {
	// ThresholdMs in milliseconds.
	ThresholdMs int64 `river:"threshold_ms,attr"`
}

func (pol PolicyCfg) convertLatencyCfg() tsp.LatencyCfg {
	otelCfg := tsp.LatencyCfg{}

	err := mapstructure.Decode(map[string]interface{}{
		"threshold_ms": pol.LatencyCfg.ThresholdMs,
	}, &otelCfg)

	if err != nil {
		panic(err)
	}

	return otelCfg
}

// NumericAttributeCfg holds the configurable settings to create a numeric attribute filter
// sampling policy evaluator.
type NumericAttributeCfg struct {
	// Tag that the filter is going to be matching against.
	Key string `river:"key,attr"`
	// MinValue is the minimum value of the attribute to be considered a match.
	MinValue int64 `river:"min_value,attr"`
	// MaxValue is the maximum value of the attribute to be considered a match.
	MaxValue int64 `river:"max_value,attr"`
}

func (pol PolicyCfg) convertNumericAttributeCfg() tsp.NumericAttributeCfg {
	otelCfg := tsp.NumericAttributeCfg{}

	err := mapstructure.Decode(map[string]interface{}{
		"key":       pol.NumericAttributeCfg.Key,
		"min_value": pol.NumericAttributeCfg.MinValue,
		"max_value": pol.NumericAttributeCfg.MaxValue,
	}, &otelCfg)

	if err != nil {
		panic(err)
	}

	return otelCfg
}
