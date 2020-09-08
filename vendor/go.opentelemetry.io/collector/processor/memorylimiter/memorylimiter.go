// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memorylimiter

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"time"

	"go.opencensus.io/stats"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/consumer/pdatautil"
	"go.opentelemetry.io/collector/internal/data"
	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/processor"
)

var (
	// errForcedDrop will be returned to callers of ConsumeTraceData to indicate
	// that data is being dropped due to high memory usage.
	errForcedDrop = errors.New("data dropped due to high memory usage")

	// Construction errors

	errNilNextConsumer = errors.New("nil nextConsumer")

	errCheckIntervalOutOfRange = errors.New(
		"checkInterval must be greater than zero")

	errMemAllocLimitOutOfRange = errors.New(
		"memAllocLimit must be greater than zero")

	errMemSpikeLimitOutOfRange = errors.New(
		"memSpikeLimit must be smaller than memAllocLimit")
)

type memoryLimiter struct {
	traceConsumer   consumer.TraceConsumer
	metricsConsumer consumer.MetricsConsumer
	logConsumer     consumer.LogConsumer

	memAllocLimit uint64
	memSpikeLimit uint64
	memCheckWait  time.Duration
	ballastSize   uint64

	// forceDrop is used atomically to indicate when data should be dropped.
	forceDrop int64

	ticker *time.Ticker

	// The function to read the mem values is set as a reference to help with
	// testing different values.
	readMemStatsFn func(m *runtime.MemStats)

	// Fields used for logging.
	procName               string
	logger                 *zap.Logger
	configMismatchedLogged bool
}

// newMemoryLimiter returns a new memorylimiter processor.
func newMemoryLimiter(
	logger *zap.Logger,
	traceConsumer consumer.TraceConsumer,
	metricsConsumer consumer.MetricsConsumer,
	logConsumer consumer.LogConsumer,
	cfg *Config) (TripleTypeProcessor, error) {
	const mibBytes = 1024 * 1024
	memAllocLimit := uint64(cfg.MemoryLimitMiB) * mibBytes
	memSpikeLimit := uint64(cfg.MemorySpikeLimitMiB) * mibBytes
	ballastSize := uint64(cfg.BallastSizeMiB) * mibBytes

	if traceConsumer == nil && metricsConsumer == nil && logConsumer == nil {
		return nil, errNilNextConsumer
	}
	if cfg.CheckInterval <= 0 {
		return nil, errCheckIntervalOutOfRange
	}
	if memAllocLimit == 0 {
		return nil, errMemAllocLimitOutOfRange
	}
	if memSpikeLimit >= memAllocLimit {
		return nil, errMemSpikeLimitOutOfRange
	}

	ml := &memoryLimiter{
		traceConsumer:   traceConsumer,
		metricsConsumer: metricsConsumer,
		logConsumer:     logConsumer,
		memAllocLimit:   memAllocLimit,
		memSpikeLimit:   memSpikeLimit,
		memCheckWait:    cfg.CheckInterval,
		ballastSize:     ballastSize,
		ticker:          time.NewTicker(cfg.CheckInterval),
		readMemStatsFn:  runtime.ReadMemStats,
		procName:        cfg.Name(),
		logger:          logger,
	}

	ml.startMonitoring()

	return ml, nil
}

func (ml *memoryLimiter) ConsumeTraces(
	ctx context.Context,
	td pdata.Traces,
) error {

	ctx = obsreport.ProcessorContext(ctx, ml.procName)
	numSpans := td.SpanCount()
	if ml.forcingDrop() {
		stats.Record(
			ctx,
			processor.StatDroppedSpanCount.M(int64(numSpans)),
			processor.StatTraceBatchesDroppedCount.M(1))

		// TODO: actually to be 100% sure that this is "refused" and not "dropped"
		// 	it is necessary to check the pipeline to see if this is directly connected
		// 	to a receiver (ie.: a receiver is on the call stack). For now it
		// 	assumes that the pipeline is properly configured and a receiver is on the
		// 	callstack.
		obsreport.ProcessorTraceDataRefused(ctx, numSpans)

		return errForcedDrop
	}

	// Even if the next consumer returns error record the data as accepted by
	// this processor.
	obsreport.ProcessorTraceDataAccepted(ctx, numSpans)
	return ml.traceConsumer.ConsumeTraces(ctx, td)
}

func (ml *memoryLimiter) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {

	ctx = obsreport.ProcessorContext(ctx, ml.procName)
	_, numDataPoints := pdatautil.MetricAndDataPointCount(md)
	if ml.forcingDrop() {
		// TODO: actually to be 100% sure that this is "refused" and not "dropped"
		// 	it is necessary to check the pipeline to see if this is directly connected
		// 	to a receiver (ie.: a receiver is on the call stack). For now it
		// 	assumes that the pipeline is properly configured and a receiver is on the
		// 	callstack.
		obsreport.ProcessorMetricsDataRefused(ctx, numDataPoints)

		return errForcedDrop
	}

	// Even if the next consumer returns error record the data as accepted by
	// this processor.
	obsreport.ProcessorMetricsDataAccepted(ctx, numDataPoints)
	return ml.metricsConsumer.ConsumeMetrics(ctx, md)
}

func (ml *memoryLimiter) ConsumeLogs(ctx context.Context, ld data.Logs) error {

	ctx = obsreport.ProcessorContext(ctx, ml.procName)
	numRecords := ld.LogRecordCount()
	if ml.forcingDrop() {
		// TODO: actually to be 100% sure that this is "refused" and not "dropped"
		// 	it is necessary to check the pipeline to see if this is directly connected
		// 	to a receiver (ie.: a receiver is on the call stack). For now it
		// 	assumes that the pipeline is properly configured and a receiver is on the
		// 	callstack.
		obsreport.ProcessorLogRecordsRefused(ctx, numRecords)

		return errForcedDrop
	}

	// Even if the next consumer returns error record the data as accepted by
	// this processor.
	obsreport.ProcessorMetricsDataAccepted(ctx, numRecords)
	return ml.logConsumer.ConsumeLogs(ctx, ld)
}

func (ml *memoryLimiter) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{MutatesConsumedData: false}
}

func (ml *memoryLimiter) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (ml *memoryLimiter) Shutdown(context.Context) error {
	ml.ticker.Stop()
	return nil
}

func (ml *memoryLimiter) readMemStats(ms *runtime.MemStats) {
	ml.readMemStatsFn(ms)
	// If proper configured ms.Alloc should be at least ml.ballastSize but since
	// a misconfiguration is possible check for that here.
	if ms.Alloc >= ml.ballastSize {
		ms.Alloc -= ml.ballastSize
	} else {
		// This indicates misconfiguration. Log it once.
		if !ml.configMismatchedLogged {
			ml.configMismatchedLogged = true
			ml.logger.Warn(typeStr + " is likely incorrectly configured. " + ballastSizeMibKey +
				" must be set equal to --mem-ballast-size-mib command line option.")
		}
	}
}

// startMonitoring starts a ticker'd goroutine that will check memory usage
// every checkInterval period.
func (ml *memoryLimiter) startMonitoring() {
	go func() {
		for range ml.ticker.C {
			ml.memCheck()
		}
	}()
}

// forcingDrop indicates when memory resources need to be released.
func (ml *memoryLimiter) forcingDrop() bool {
	return atomic.LoadInt64(&ml.forceDrop) != 0
}

func (ml *memoryLimiter) memCheck() {
	ms := &runtime.MemStats{}
	ml.readMemStats(ms)
	ml.memLimiting(ms)
}

func (ml *memoryLimiter) shouldForceDrop(ms *runtime.MemStats) bool {
	return ml.memAllocLimit <= ms.Alloc || ml.memAllocLimit-ms.Alloc <= ml.memSpikeLimit
}

func (ml *memoryLimiter) memLimiting(ms *runtime.MemStats) {
	if !ml.shouldForceDrop(ms) {
		atomic.StoreInt64(&ml.forceDrop, 0)
	} else {
		atomic.StoreInt64(&ml.forceDrop, 1)
		// Force a GC at this point and see if this is enough to get to
		// the desired level.
		runtime.GC()
	}
}
