// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tailsamplingprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"

import (
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor/internal/sampling"
)

func getNewCompositePolicy(logger *zap.Logger, config *CompositeCfg) (sampling.PolicyEvaluator, error) {
	var subPolicyEvalParams []sampling.SubPolicyEvalParams
	rateAllocationsMap := getRateAllocationMap(config)
	for i := range config.SubPolicyCfg {
		policyCfg := &config.SubPolicyCfg[i]
		policy, err := getCompositeSubPolicyEvaluator(logger, policyCfg)
		if err != nil {
			return nil, err
		}

		evalParams := sampling.SubPolicyEvalParams{
			Evaluator:         policy,
			MaxSpansPerSecond: int64(rateAllocationsMap[policyCfg.Name]),
		}
		subPolicyEvalParams = append(subPolicyEvalParams, evalParams)
	}
	return sampling.NewComposite(logger, config.MaxTotalSpansPerSecond, subPolicyEvalParams, sampling.MonotonicClock{}), nil
}

// Apply rate allocations to the sub-policies
func getRateAllocationMap(config *CompositeCfg) map[string]float64 {
	rateAllocationsMap := make(map[string]float64)
	maxTotalSPS := float64(config.MaxTotalSpansPerSecond)
	// Default SPS determined by equally diving number of sub policies
	defaultSPS := maxTotalSPS / float64(len(config.SubPolicyCfg))
	for _, rAlloc := range config.RateAllocation {
		if rAlloc.Percent > 0 {
			rateAllocationsMap[rAlloc.Policy] = (float64(rAlloc.Percent) / 100) * maxTotalSPS
		} else {
			rateAllocationsMap[rAlloc.Policy] = defaultSPS
		}
	}
	return rateAllocationsMap
}

// Return instance of composite sub-policy
func getCompositeSubPolicyEvaluator(logger *zap.Logger, cfg *CompositeSubPolicyCfg) (sampling.PolicyEvaluator, error) {
	switch cfg.Type {
	case And:
		return getNewAndPolicy(logger, &cfg.AndCfg)
	default:
		return getSharedPolicyEvaluator(logger, &cfg.sharedPolicyCfg)
	}
}
