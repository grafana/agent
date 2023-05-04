// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tailsamplingprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/timeutils"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor/internal/idbatcher"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor/internal/sampling"
)

// policy combines a sampling policy evaluator with the destinations to be
// used for that policy.
type policy struct {
	// name used to identify this policy instance.
	name string
	// evaluator that decides if a trace is sampled or not by this policy instance.
	evaluator sampling.PolicyEvaluator
	// ctx used to carry metric tags of each policy.
	ctx context.Context
}

// tailSamplingSpanProcessor handles the incoming trace data and uses the given sampling
// policy to sample traces.
type tailSamplingSpanProcessor struct {
	ctx             context.Context
	nextConsumer    consumer.Traces
	maxNumTraces    uint64
	policies        []*policy
	logger          *zap.Logger
	idToTrace       sync.Map
	policyTicker    timeutils.TTicker
	tickerFrequency time.Duration
	decisionBatcher idbatcher.Batcher
	deleteChan      chan pcommon.TraceID
	numTracesOnMap  *atomic.Uint64
}

const (
	sourceFormat = "tail_sampling"
)

// newTracesProcessor returns a processor.TracesProcessor that will perform tail sampling according to the given
// configuration.
func newTracesProcessor(logger *zap.Logger, nextConsumer consumer.Traces, cfg Config) (component.TracesProcessor, error) {
	if nextConsumer == nil {
		return nil, component.ErrNilNextConsumer
	}

	numDecisionBatches := uint64(cfg.DecisionWait.Seconds())
	inBatcher, err := idbatcher.New(numDecisionBatches, cfg.ExpectedNewTracesPerSec, uint64(2*runtime.NumCPU()))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var policies []*policy
	for i := range cfg.PolicyCfgs {
		policyCfg := &cfg.PolicyCfgs[i]
		policyCtx, err := tag.New(ctx, tag.Upsert(tagPolicyKey, policyCfg.Name), tag.Upsert(tagSourceFormat, sourceFormat))
		if err != nil {
			return nil, err
		}
		eval, err := getPolicyEvaluator(logger, policyCfg)
		if err != nil {
			return nil, err
		}
		p := &policy{
			name:      policyCfg.Name,
			evaluator: eval,
			ctx:       policyCtx,
		}
		policies = append(policies, p)
	}

	tsp := &tailSamplingSpanProcessor{
		ctx:             ctx,
		nextConsumer:    nextConsumer,
		maxNumTraces:    cfg.NumTraces,
		logger:          logger,
		decisionBatcher: inBatcher,
		policies:        policies,
		tickerFrequency: time.Second,
		numTracesOnMap:  atomic.NewUint64(0),
	}

	tsp.policyTicker = &timeutils.PolicyTicker{OnTickFunc: tsp.samplingPolicyOnTick}
	tsp.deleteChan = make(chan pcommon.TraceID, cfg.NumTraces)

	return tsp, nil
}

func getPolicyEvaluator(logger *zap.Logger, cfg *PolicyCfg) (sampling.PolicyEvaluator, error) {
	switch cfg.Type {
	case Composite:
		return getNewCompositePolicy(logger, &cfg.CompositeCfg)
	case And:
		return getNewAndPolicy(logger, &cfg.AndCfg)
	default:
		return getSharedPolicyEvaluator(logger, &cfg.sharedPolicyCfg)
	}
}

func getSharedPolicyEvaluator(logger *zap.Logger, cfg *sharedPolicyCfg) (sampling.PolicyEvaluator, error) {
	switch cfg.Type {
	case AlwaysSample:
		return sampling.NewAlwaysSample(logger), nil
	case Latency:
		lfCfg := cfg.LatencyCfg
		return sampling.NewLatency(logger, lfCfg.ThresholdMs), nil
	case NumericAttribute:
		nafCfg := cfg.NumericAttributeCfg
		return sampling.NewNumericAttributeFilter(logger, nafCfg.Key, nafCfg.MinValue, nafCfg.MaxValue), nil
	case Probabilistic:
		pCfg := cfg.ProbabilisticCfg
		return sampling.NewProbabilisticSampler(logger, pCfg.HashSalt, pCfg.SamplingPercentage), nil
	case StringAttribute:
		safCfg := cfg.StringAttributeCfg
		return sampling.NewStringAttributeFilter(logger, safCfg.Key, safCfg.Values, safCfg.EnabledRegexMatching, safCfg.CacheMaxSize, safCfg.InvertMatch), nil
	case StatusCode:
		scfCfg := cfg.StatusCodeCfg
		return sampling.NewStatusCodeFilter(logger, scfCfg.StatusCodes)
	case RateLimiting:
		rlfCfg := cfg.RateLimitingCfg
		return sampling.NewRateLimiting(logger, rlfCfg.SpansPerSecond), nil
	case SpanCount:
		spCfg := cfg.SpanCountCfg
		return sampling.NewSpanCount(logger, spCfg.MinSpans), nil
	case TraceState:
		tsfCfg := cfg.TraceStateCfg
		return sampling.NewTraceStateFilter(logger, tsfCfg.Key, tsfCfg.Values), nil
	default:
		return nil, fmt.Errorf("unknown sampling policy type %s", cfg.Type)
	}
}

type policyMetrics struct {
	idNotFoundOnMapCount, evaluateErrorCount, decisionSampled, decisionNotSampled int64
}

func (tsp *tailSamplingSpanProcessor) samplingPolicyOnTick() {
	metrics := policyMetrics{}

	startTime := time.Now()
	batch, _ := tsp.decisionBatcher.CloseCurrentAndTakeFirstBatch()
	batchLen := len(batch)
	tsp.logger.Debug("Sampling Policy Evaluation ticked")
	for _, id := range batch {
		d, ok := tsp.idToTrace.Load(id)
		if !ok {
			metrics.idNotFoundOnMapCount++
			continue
		}
		trace := d.(*sampling.TraceData)
		trace.DecisionTime = time.Now()

		decision, policy := tsp.makeDecision(id, trace, &metrics)

		// Sampled or not, remove the batches
		trace.Lock()
		allSpans := ptrace.NewTraces()
		trace.ReceivedBatches.MoveTo(allSpans)
		trace.Unlock()

		if decision == sampling.Sampled {
			_ = tsp.nextConsumer.ConsumeTraces(policy.ctx, allSpans)
		}
	}

	stats.Record(tsp.ctx,
		statOverallDecisionLatencyUs.M(int64(time.Since(startTime)/time.Microsecond)),
		statDroppedTooEarlyCount.M(metrics.idNotFoundOnMapCount),
		statPolicyEvaluationErrorCount.M(metrics.evaluateErrorCount),
		statTracesOnMemoryGauge.M(int64(tsp.numTracesOnMap.Load())))

	tsp.logger.Debug("Sampling policy evaluation completed",
		zap.Int("batch.len", batchLen),
		zap.Int64("sampled", metrics.decisionSampled),
		zap.Int64("notSampled", metrics.decisionNotSampled),
		zap.Int64("droppedPriorToEvaluation", metrics.idNotFoundOnMapCount),
		zap.Int64("policyEvaluationErrors", metrics.evaluateErrorCount),
	)
}

func (tsp *tailSamplingSpanProcessor) makeDecision(id pcommon.TraceID, trace *sampling.TraceData, metrics *policyMetrics) (sampling.Decision, *policy) {
	finalDecision := sampling.NotSampled
	var matchingPolicy *policy
	samplingDecision := map[sampling.Decision]bool{
		sampling.Error:            false,
		sampling.Sampled:          false,
		sampling.NotSampled:       false,
		sampling.InvertSampled:    false,
		sampling.InvertNotSampled: false,
	}

	// Check all policies before making a final decision
	for i, p := range tsp.policies {
		policyEvaluateStartTime := time.Now()
		decision, err := p.evaluator.Evaluate(id, trace)
		stats.Record(
			p.ctx,
			statDecisionLatencyMicroSec.M(int64(time.Since(policyEvaluateStartTime)/time.Microsecond)))

		if err != nil {
			samplingDecision[sampling.Error] = true
			trace.Decisions[i] = sampling.NotSampled
			metrics.evaluateErrorCount++
			tsp.logger.Debug("Sampling policy error", zap.Error(err))
		} else {
			switch decision {
			case sampling.Sampled:
				samplingDecision[sampling.Sampled] = true
				trace.Decisions[i] = decision

			case sampling.NotSampled:
				samplingDecision[sampling.NotSampled] = true
				trace.Decisions[i] = decision

			case sampling.InvertSampled:
				samplingDecision[sampling.InvertSampled] = true
				trace.Decisions[i] = sampling.Sampled

			case sampling.InvertNotSampled:
				samplingDecision[sampling.InvertNotSampled] = true
				trace.Decisions[i] = sampling.NotSampled
			}
		}
	}

	// InvertNotSampled takes precedence over any other decision
	switch {
	case samplingDecision[sampling.InvertNotSampled]:
		finalDecision = sampling.NotSampled
	case samplingDecision[sampling.Sampled]:
		finalDecision = sampling.Sampled
	case samplingDecision[sampling.InvertSampled] && !samplingDecision[sampling.NotSampled]:
		finalDecision = sampling.Sampled
	}

	for _, p := range tsp.policies {
		switch finalDecision {
		case sampling.Sampled:
			// any single policy that decides to sample will cause the decision to be sampled
			// the nextConsumer will get the context from the first matching policy
			if matchingPolicy == nil {
				matchingPolicy = p
			}

			_ = stats.RecordWithTags(
				p.ctx,
				[]tag.Mutator{tag.Upsert(tagSampledKey, "true")},
				statCountTracesSampled.M(int64(1)),
			)
			metrics.decisionSampled++

		case sampling.NotSampled:
			_ = stats.RecordWithTags(
				p.ctx,
				[]tag.Mutator{tag.Upsert(tagSampledKey, "false")},
				statCountTracesSampled.M(int64(1)),
			)
			metrics.decisionNotSampled++
		}
	}

	return finalDecision, matchingPolicy
}

// ConsumeTraces is required by the component.TracesProcessor interface.
func (tsp *tailSamplingSpanProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	resourceSpans := td.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		tsp.processTraces(resourceSpans.At(i))
	}
	return nil
}

func (tsp *tailSamplingSpanProcessor) groupSpansByTraceKey(resourceSpans ptrace.ResourceSpans) map[pcommon.TraceID][]*ptrace.Span {
	idToSpans := make(map[pcommon.TraceID][]*ptrace.Span)
	ilss := resourceSpans.ScopeSpans()
	for j := 0; j < ilss.Len(); j++ {
		spans := ilss.At(j).Spans()
		spansLen := spans.Len()
		for k := 0; k < spansLen; k++ {
			span := spans.At(k)
			key := span.TraceID()
			idToSpans[key] = append(idToSpans[key], &span)
		}
	}
	return idToSpans
}

func (tsp *tailSamplingSpanProcessor) processTraces(resourceSpans ptrace.ResourceSpans) {
	// Group spans per their traceId to minimize contention on idToTrace
	idToSpans := tsp.groupSpansByTraceKey(resourceSpans)
	var newTraceIDs int64
	for id, spans := range idToSpans {
		lenSpans := int64(len(spans))
		lenPolicies := len(tsp.policies)
		initialDecisions := make([]sampling.Decision, lenPolicies)
		for i := 0; i < lenPolicies; i++ {
			initialDecisions[i] = sampling.Pending
		}
		d, loaded := tsp.idToTrace.Load(id)
		if !loaded {
			d, loaded = tsp.idToTrace.LoadOrStore(id, &sampling.TraceData{
				Decisions:       initialDecisions,
				ArrivalTime:     time.Now(),
				SpanCount:       atomic.NewInt64(lenSpans),
				ReceivedBatches: ptrace.NewTraces(),
			})
		}
		actualData := d.(*sampling.TraceData)
		if loaded {
			actualData.SpanCount.Add(lenSpans)
		} else {
			newTraceIDs++
			tsp.decisionBatcher.AddToCurrentBatch(id)
			tsp.numTracesOnMap.Add(1)
			postDeletion := false
			currTime := time.Now()
			for !postDeletion {
				select {
				case tsp.deleteChan <- id:
					postDeletion = true
				default:
					traceKeyToDrop := <-tsp.deleteChan
					tsp.dropTrace(traceKeyToDrop, currTime)
				}
			}
		}

		for i, p := range tsp.policies {
			actualData.Lock()
			actualDecision := actualData.Decisions[i]
			// If decision is pending, we want to add the new spans still under the lock, so the decision doesn't happen
			// in between the transition from pending.
			if actualDecision == sampling.Pending {
				// Add the spans to the trace, but only once for all policy, otherwise same spans will
				// be duplicated in the final trace.
				appendToTraces(actualData.ReceivedBatches, resourceSpans, spans)
				actualData.Unlock()
				break
			}
			actualData.Unlock()

			switch actualDecision {
			case sampling.Sampled:
				// Forward the spans to the policy destinations
				traceTd := ptrace.NewTraces()
				appendToTraces(traceTd, resourceSpans, spans)
				if err := tsp.nextConsumer.ConsumeTraces(p.ctx, traceTd); err != nil {
					tsp.logger.Warn("Error sending late arrived spans to destination",
						zap.String("policy", p.name),
						zap.Error(err))
				}
			case sampling.NotSampled:
				stats.Record(tsp.ctx, statLateSpanArrivalAfterDecision.M(int64(time.Since(actualData.DecisionTime)/time.Second)))

			default:
				tsp.logger.Warn("Encountered unexpected sampling decision",
					zap.String("policy", p.name),
					zap.Int("decision", int(actualDecision)))
			}

			// At this point the late arrival has been passed to nextConsumer. Need to break out of the policy loop
			// so that it isn't sent to nextConsumer more than once when multiple policies chose to sample
			if actualDecision == sampling.Sampled {
				break
			}
		}
	}

	stats.Record(tsp.ctx, statNewTraceIDReceivedCount.M(newTraceIDs))
}

func (tsp *tailSamplingSpanProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// Start is invoked during service startup.
func (tsp *tailSamplingSpanProcessor) Start(context.Context, component.Host) error {
	tsp.policyTicker.Start(tsp.tickerFrequency)
	return nil
}

// Shutdown is invoked during service shutdown.
func (tsp *tailSamplingSpanProcessor) Shutdown(context.Context) error {
	tsp.decisionBatcher.Stop()
	tsp.policyTicker.Stop()
	return nil
}

func (tsp *tailSamplingSpanProcessor) dropTrace(traceID pcommon.TraceID, deletionTime time.Time) {
	var trace *sampling.TraceData
	if d, ok := tsp.idToTrace.Load(traceID); ok {
		trace = d.(*sampling.TraceData)
		tsp.idToTrace.Delete(traceID)
		// Subtract one from numTracesOnMap per https://godoc.org/sync/atomic#AddUint64
		tsp.numTracesOnMap.Add(^uint64(0))
	}
	if trace == nil {
		tsp.logger.Error("Attempt to delete traceID not on table")
		return
	}

	stats.Record(tsp.ctx, statTraceRemovalAgeSec.M(int64(deletionTime.Sub(trace.ArrivalTime)/time.Second)))
}

func appendToTraces(dest ptrace.Traces, rss ptrace.ResourceSpans, spans []*ptrace.Span) {
	rs := dest.ResourceSpans().AppendEmpty()
	rss.Resource().CopyTo(rs.Resource())
	ils := rs.ScopeSpans().AppendEmpty()
	for _, span := range spans {
		sp := ils.Spans().AppendEmpty()
		span.CopyTo(sp)
	}
}
