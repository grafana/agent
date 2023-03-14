// Copyright (c) 2018 The Jaeger Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strategy_store

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/jaegertracing/jaeger/thrift-gen/sampling"
)

// strategiesJSON returns the strategy with
// a given probability.
func strategiesJSON(probability float32) string {
	strategy := fmt.Sprintf(`
		{
			"default_strategy": {
			"type": "probabilistic",
			"param": 0.5
			},
			"service_strategies": [
			{
				"service": "foo",
				"type": "probabilistic",
				"param": %.1f
			},
			{
				"service": "bar",
				"type": "ratelimiting",
				"param": 5
			},
			{
				"service": "foo-per-op",
				"type": "probabilistic",
				"param": 0.8,
				"operation_strategies": [
					{
					"operation": "op1",
					"type": "probabilistic",
					"param": 0.2
					},
					{
					"operation": "op2",
					"type": "probabilistic",
					"param": 0.4
					}
				]
			}
			]
		}
		`,
		probability,
	)
	return strategy
}

func TestStrategyStore(t *testing.T) {
	store, err := NewStrategyStore(strategiesJSON(.8), zap.NewNop())
	require.NoError(t, err)
	s, err := store.GetSamplingStrategy(context.Background(), "foo")
	require.NoError(t, err)
	assert.EqualValues(t, makeResponse(sampling.SamplingStrategyType_PROBABILISTIC, 0.8), *s)

	s, err = store.GetSamplingStrategy(context.Background(), "bar")
	require.NoError(t, err)
	assert.EqualValues(t, makeResponse(sampling.SamplingStrategyType_RATE_LIMITING, 5), *s)

	s, err = store.GetSamplingStrategy(context.Background(), "default")
	require.NoError(t, err)
	assert.EqualValues(t, makeResponse(sampling.SamplingStrategyType_PROBABILISTIC, 0.5), *s)

	s, err = store.GetSamplingStrategy(context.Background(), "foo-per-op")
	require.NoError(t, err)
	expected := makeResponse(sampling.SamplingStrategyType_PROBABILISTIC, 0.8)
	expected.OperationSampling = &sampling.PerOperationSamplingStrategies{
		DefaultSamplingProbability:       0.8,
		DefaultLowerBoundTracesPerSecond: 0,
		PerOperationStrategies: []*sampling.OperationSamplingStrategy{
			{
				Operation: "op1",
				ProbabilisticSampling: &sampling.ProbabilisticSamplingStrategy{
					SamplingRate: 0.2,
				},
			},
			{
				Operation: "op2",
				ProbabilisticSampling: &sampling.ProbabilisticSamplingStrategy{
					SamplingRate: 0.4,
				},
			},
		},
	}
	assert.EqualValues(t, expected, *s)

}

func makeResponse(samplerType sampling.SamplingStrategyType, param float64) (resp sampling.SamplingStrategyResponse) {
	resp.StrategyType = samplerType
	if samplerType == sampling.SamplingStrategyType_PROBABILISTIC {
		resp.ProbabilisticSampling = &sampling.ProbabilisticSamplingStrategy{
			SamplingRate: param,
		}
	} else if samplerType == sampling.SamplingStrategyType_RATE_LIMITING {
		resp.RateLimitingSampling = &sampling.RateLimitingSamplingStrategy{
			MaxTracesPerSecond: int16(param),
		}
	}
	return resp
}
