package stages

// This package is ported over from grafana/loki/clients/pkg/logentry/stages.
// We aim to port the stages in steps, to avoid introducing huge amounts of
// new code without being able to slowly review, examine and test them.

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

// StageConfig defines a single stage in a processing pipeline.
// We define these as pointers types so we can use reflection to check that
// exactly one is set.
type StageConfig struct {
	JSONConfig   *JSONConfig   `river:"json,block,optional"`
	LabelsConfig *LabelsConfig `river:"labels,block,optional"`
}

// UnmarshalRiver implements river.Unmarshaler.
func (arg *StageConfig) UnmarshalRiver(f func(interface{}) error) error {
	// *arg = DefaultArguments
	type args StageConfig
	if err := f((*args)(arg)); err != nil {
		return err
	}

	nonEmpty := 0
	v := reflect.Indirect(reflect.ValueOf(arg))
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).IsNil() {
			nonEmpty++
		}
	}

	if nonEmpty != 1 {
		return fmt.Errorf("each stage block should contain exactly one stage definition, found %d", nonEmpty)
	}

	return nil
}

var rateLimiter *rate.Limiter
var rateLimiterDrop bool
var rateLimiterDropReason = "global_rate_limiter_drop"

// Pipeline pass down a log entry to each stage for mutation and/or label extraction.
type Pipeline struct {
	logger    log.Logger
	stages    []Stage
	jobName   *string
	dropCount *prometheus.CounterVec
}

// NewPipeline creates a new log entry pipeline from a configuration
func NewPipeline(logger log.Logger, stages []StageConfig, jobName *string, registerer prometheus.Registerer) (*Pipeline, error) {
	st := []Stage{}
	for _, stage := range stages {
		newStage, err := New(logger, jobName, stage, registerer)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid stage config")
		}
		st = append(st, newStage)
	}
	return &Pipeline{
		logger:    log.With(logger, "component", "pipeline"),
		stages:    st,
		jobName:   jobName,
		dropCount: getDropCountMetric(registerer),
	}, nil
}

// RunWith will reads from the input channel entries, mutate them with the process function and returns them via the output channel.
func RunWith(input chan Entry, process func(e Entry) Entry) chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)
		for e := range input {
			out <- process(e)
		}
	}()
	return out
}

// RunWithSkip same as RunWith, except it skip sending it to output channel, if `process` functions returns `skip` true.
func RunWithSkip(input chan Entry, process func(e Entry) (Entry, bool)) chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)
		for e := range input {
			ee, skip := process(e)
			if skip {
				continue
			}
			out <- ee
		}
	}()

	return out
}

// Run implements Stage
func (p *Pipeline) Run(in chan Entry) chan Entry {
	in = RunWith(in, func(e Entry) Entry {
		// Initialize the extracted map with the initial labels (ie. "filename"),
		// so that stages can operate on initial labels too
		for labelName, labelValue := range e.Labels {
			e.Extracted[string(labelName)] = string(labelValue)
		}
		return e
	})
	// chain all stages together.
	for _, m := range p.stages {
		in = m.Run(in)
	}
	return in
}

// Name implements Stage
func (p *Pipeline) Name() string {
	return StageTypePipeline
}

// Wrap implements EntryMiddleware
func (p *Pipeline) Wrap(next loki.EntryHandler) loki.EntryHandler {
	handlerIn := make(chan loki.Entry)
	nextChan := next.Chan()
	wg, once := sync.WaitGroup{}, sync.Once{}
	pipelineIn := make(chan Entry)
	pipelineOut := p.Run(pipelineIn)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for e := range pipelineOut {
			if rateLimiter != nil {
				if rateLimiterDrop {
					if !rateLimiter.Allow() {
						p.dropCount.WithLabelValues(rateLimiterDropReason).Inc()
						continue
					}
				} else {
					_ = rateLimiter.Wait(context.Background())
				}
			}
			nextChan <- e.Entry
		}
	}()
	go func() {
		defer wg.Done()
		defer close(pipelineIn)
		for e := range handlerIn {
			pipelineIn <- Entry{
				Extracted: map[string]interface{}{},
				Entry:     e,
			}
		}
	}()
	return loki.NewEntryHandler(handlerIn, func() {
		once.Do(func() { close(handlerIn) })
		wg.Wait()
	})
}

// Size gets the current number of stages in the pipeline
func (p *Pipeline) Size() int {
	return len(p.stages)
}

func SetReadLineRateLimiter(rateVal float64, burstVal int, drop bool) {
	rateLimiter = rate.NewLimiter(rate.Limit(rateVal), burstVal)
	rateLimiterDrop = drop
}

// TODO(@tpaschalis) This is a helper from the metrics stage. Remove this
// copy when we port it over, or remove it entirely in favor of a central
// metrics struct.
func getDropCountMetric(registerer prometheus.Registerer) *prometheus.CounterVec {
	dropCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "loki_process_dropped_lines_total",
		Help: "A count of all log lines dropped as a result of a pipeline stage",
	}, []string{"reason"})
	err := registerer.Register(dropCount)
	if err != nil {
		if existing, ok := err.(prometheus.AlreadyRegisteredError); ok {
			dropCount = existing.ExistingCollector.(*prometheus.CounterVec)
		} else {
			// Same behavior as MustRegister if the error is not for AlreadyRegistered
			panic(err)
		}
	}
	return dropCount
}
