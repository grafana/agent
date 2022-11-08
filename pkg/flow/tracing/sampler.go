package tracing

import (
	"sync"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type dynamicSampler struct {
	mut   sync.RWMutex
	inner tracesdk.Sampler
}

var _ tracesdk.Sampler = (*dynamicSampler)(nil)

func newDynamicSampler(rate float64) *dynamicSampler {
	var ds dynamicSampler
	ds.UpdateSampleRate(rate)
	return &ds
}

func (ds *dynamicSampler) UpdateSampleRate(rate float64) {
	ds.mut.Lock()
	defer ds.mut.Unlock()

	ds.inner = tracesdk.TraceIDRatioBased(rate)
}

func (ds *dynamicSampler) ShouldSample(parameters tracesdk.SamplingParameters) tracesdk.SamplingResult {
	ds.mut.RLock()
	defer ds.mut.RUnlock()
	return ds.inner.ShouldSample(parameters)
}

func (ds *dynamicSampler) Description() string {
	ds.mut.RLock()
	defer ds.mut.RUnlock()
	return ds.inner.Description()
}
