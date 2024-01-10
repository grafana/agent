package processortest

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/internal/fakeconsumer"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/flow/logging/level"
	"github.com/grafana/dskit/backoff"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

//
// Utilities for running a processor end-to-end and testing its outputs.
// They work for metrics, logs and traces.
//

type Signal interface {
	MakeOutput() *otelcol.ConsumerArguments
	ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error
	CheckOutput(t *testing.T)
}

type ProcessorRunConfig struct {
	Ctx        context.Context
	T          *testing.T
	Args       component.Arguments
	TestSignal Signal
	Ctrl       *componenttest.Controller
	L          log.Logger
}

func TestRunProcessor(c ProcessorRunConfig) {
	go func() {
		err := c.Ctrl.Run(c.Ctx, c.Args)
		require.NoError(c.T, err)
	}()

	require.NoError(c.T, c.Ctrl.WaitRunning(time.Second), "component never started")
	require.NoError(c.T, c.Ctrl.WaitExports(time.Second), "component never exported anything")

	// Send signals in the background to our processor.
	go func() {
		exports := c.Ctrl.Exports().(otelcol.ConsumerExports)

		bo := backoff.New(c.Ctx, backoff.Config{
			MinBackoff: 10 * time.Millisecond,
			MaxBackoff: 100 * time.Millisecond,
		})
		for bo.Ongoing() {
			err := c.TestSignal.ConsumeInput(c.Ctx, exports.Input)
			if err != nil {
				level.Error(c.L).Log("msg", "failed to send signal", "err", err)
				bo.Wait()
				continue
			}

			return
		}
	}()

	c.TestSignal.CheckOutput(c.T)
}

//
// Trace to Logs
//

type traceToLogSignal struct {
	logCh              chan plog.Logs
	inputTrace         ptrace.Traces
	expectedOuutputLog plog.Logs
}

func NewTraceToLogSignal(inputJson string, expectedOutputJson string) Signal {
	return &traceToLogSignal{
		logCh:              make(chan plog.Logs),
		inputTrace:         CreateTestTraces(inputJson),
		expectedOuutputLog: CreateTestLogs(expectedOutputJson),
	}
}

func (s traceToLogSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeLogsOutput(s.logCh)
}

func (s traceToLogSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeTraces(ctx, s.inputTrace)
}

func (s traceToLogSignal) CheckOutput(t *testing.T) {
	// Wait for our processor to finish and forward data to logCh.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for logs")
	case tr := <-s.logCh:
		trStr := marshalLogs(tr)
		expStr := marshalLogs(s.expectedOuutputLog)
		require.JSONEq(t, expStr, trStr)
	}
}

//
// Trace to Metrics
//

type traceToMetricSignal struct {
	metricCh              chan pmetric.Metrics
	inputTrace            ptrace.Traces
	expectedOuutputMetric pmetric.Metrics
}

// Any timestamps inside expectedOutputJson should be set to 0.
func NewTraceToMetricSignal(inputJson string, expectedOutputJson string) Signal {
	return &traceToMetricSignal{
		metricCh:              make(chan pmetric.Metrics),
		inputTrace:            CreateTestTraces(inputJson),
		expectedOuutputMetric: CreateTestMetrics(expectedOutputJson),
	}
}

func (s traceToMetricSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeMetricsOutput(s.metricCh)
}

func (s traceToMetricSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeTraces(ctx, s.inputTrace)
}

// Set the timestamp of all data points to 0.
// This helps avoid flaky tests due to timestamps.
func setMetricTimestampToZero(metrics pmetric.Metrics) {
	// Loop over all resource metrics
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		// Loop over all metric scopes.
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			// Loop over all metrics.
			for k := 0; k < sm.Metrics().Len(); k++ {
				m := sm.Metrics().At(k)
				switch m.Type() {
				case pmetric.MetricTypeSum:
					// Loop over all data points.
					for l := 0; l < m.Sum().DataPoints().Len(); l++ {
						// Set the timestamp to 0 to avoid flaky tests.
						dp := m.Sum().DataPoints().At(l)
						dp.SetTimestamp(0)
						dp.SetStartTimestamp(0)
					}
				case pmetric.MetricTypeGauge:
					// Loop over all data points.
					for l := 0; l < m.Gauge().DataPoints().Len(); l++ {
						// Set the timestamp to 0 to avoid flaky tests.
						dp := m.Gauge().DataPoints().At(l)
						dp.SetTimestamp(0)
						dp.SetStartTimestamp(0)
					}
				case pmetric.MetricTypeHistogram:
					// Loop over all data points.
					for l := 0; l < m.Histogram().DataPoints().Len(); l++ {
						// Set the timestamp to 0 to avoid flaky tests.
						dp := m.Histogram().DataPoints().At(l)
						dp.SetTimestamp(0)
						dp.SetStartTimestamp(0)
					}
				case pmetric.MetricTypeSummary:
					// Loop over all data points.
					for l := 0; l < m.Summary().DataPoints().Len(); l++ {
						// Set the timestamp to 0 to avoid flaky tests.
						dp := m.Summary().DataPoints().At(l)
						dp.SetTimestamp(0)
						dp.SetStartTimestamp(0)
					}
				}
			}
		}
	}
}

// Wait for the component to finish and check its output.
func (s traceToMetricSignal) CheckOutput(t *testing.T) {
	// Set the timeout to a few seconds so that all components have finished.
	// Components such as otelcol.connector.spanmetrics may need a few
	// seconds before they output metrics.
	timeout := time.Second * 5

	select {
	case <-time.After(timeout):
		require.FailNow(t, "failed waiting for metrics")
	case tr := <-s.metricCh:
		setMetricTimestampToZero(tr)
		trStr := marshalMetrics(tr)

		expStr := marshalMetrics(s.expectedOuutputMetric)
		// Set a field from the json to an empty string to avoid flaky tests containing timestamps.

		require.JSONEq(t, expStr, trStr)
	}
}

//
// Traces
//

type traceSignal struct {
	traceCh              chan ptrace.Traces
	inputTrace           ptrace.Traces
	expectedOuutputTrace ptrace.Traces
}

func NewTraceSignal(inputJson string, expectedOutputJson string) Signal {
	return &traceSignal{
		traceCh:              make(chan ptrace.Traces),
		inputTrace:           CreateTestTraces(inputJson),
		expectedOuutputTrace: CreateTestTraces(expectedOutputJson),
	}
}

func (s traceSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeTracesOutput(s.traceCh)
}

func (s traceSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeTraces(ctx, s.inputTrace)
}

func (s traceSignal) CheckOutput(t *testing.T) {
	// Wait for our processor to finish and forward data to traceCh.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for traces")
	case tr := <-s.traceCh:
		trStr := marshalTraces(tr)
		expStr := marshalTraces(s.expectedOuutputTrace)
		require.JSONEq(t, expStr, trStr)
	}
}

// traceJson should match format from the protobuf definition:
// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/trace/v1/trace.proto
func CreateTestTraces(traceJson string) ptrace.Traces {
	decoder := &ptrace.JSONUnmarshaler{}
	data, err := decoder.UnmarshalTraces([]byte(traceJson))
	if err != nil {
		panic(err)
	}
	return data
}

func marshalTraces(trace ptrace.Traces) string {
	marshaler := &ptrace.JSONMarshaler{}
	data, err := marshaler.MarshalTraces(trace)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// makeTracesOutput returns ConsumerArguments which will forward traces to the
// provided channel.
func makeTracesOutput(ch chan ptrace.Traces) *otelcol.ConsumerArguments {
	traceConsumer := fakeconsumer.Consumer{
		ConsumeTracesFunc: func(ctx context.Context, t ptrace.Traces) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Traces: []otelcol.Consumer{&traceConsumer},
	}
}

//
// Logs
//

type logSignal struct {
	logCh              chan plog.Logs
	inputLog           plog.Logs
	expectedOuutputLog plog.Logs
}

func NewLogSignal(inputJson string, expectedOutputJson string) Signal {
	return &logSignal{
		logCh:              make(chan plog.Logs),
		inputLog:           CreateTestLogs(inputJson),
		expectedOuutputLog: CreateTestLogs(expectedOutputJson),
	}
}

func (s logSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeLogsOutput(s.logCh)
}

func (s logSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeLogs(ctx, s.inputLog)
}

func (s logSignal) CheckOutput(t *testing.T) {
	// Wait for our processor to finish and forward data to logCh.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for logs")
	case tr := <-s.logCh:
		trStr := marshalLogs(tr)
		expStr := marshalLogs(s.expectedOuutputLog)
		require.JSONEq(t, expStr, trStr)
	}
}

// makeLogsOutput returns ConsumerArguments which will forward logs to the
// provided channel.
func makeLogsOutput(ch chan plog.Logs) *otelcol.ConsumerArguments {
	logConsumer := fakeconsumer.Consumer{
		ConsumeLogsFunc: func(ctx context.Context, t plog.Logs) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Logs: []otelcol.Consumer{&logConsumer},
	}
}

// logJson should match format from the protobuf definition:
// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto
func CreateTestLogs(logJson string) plog.Logs {
	decoder := &plog.JSONUnmarshaler{}
	data, err := decoder.UnmarshalLogs([]byte(logJson))
	if err != nil {
		panic(err)
	}
	return data
}

func marshalLogs(log plog.Logs) string {
	marshaler := &plog.JSONMarshaler{}
	data, err := marshaler.MarshalLogs(log)
	if err != nil {
		panic(err)
	}
	return string(data)
}

//
// Metrics
//

type metricSignal struct {
	metricCh              chan pmetric.Metrics
	inputMetric           pmetric.Metrics
	expectedOuutputMetric pmetric.Metrics
}

func NewMetricSignal(inputJson string, expectedOutputJson string) Signal {
	return &metricSignal{
		metricCh:              make(chan pmetric.Metrics),
		inputMetric:           CreateTestMetrics(inputJson),
		expectedOuutputMetric: CreateTestMetrics(expectedOutputJson),
	}
}

func (s metricSignal) MakeOutput() *otelcol.ConsumerArguments {
	return makeMetricsOutput(s.metricCh)
}

func (s metricSignal) ConsumeInput(ctx context.Context, consumer otelcol.Consumer) error {
	return consumer.ConsumeMetrics(ctx, s.inputMetric)
}

func (s metricSignal) CheckOutput(t *testing.T) {
	// Wait for our processor to finish and forward data to logCh.
	select {
	case <-time.After(time.Second):
		require.FailNow(t, "failed waiting for logs")
	case tr := <-s.metricCh:
		trStr := marshalMetrics(tr)
		expStr := marshalMetrics(s.expectedOuutputMetric)
		require.JSONEq(t, expStr, trStr)
	}
}

// makeMetricsOutput returns ConsumerArguments which will forward metrics to the
// provided channel.
func makeMetricsOutput(ch chan pmetric.Metrics) *otelcol.ConsumerArguments {
	metricConsumer := fakeconsumer.Consumer{
		ConsumeMetricsFunc: func(ctx context.Context, t pmetric.Metrics) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- t:
				return nil
			}
		},
	}

	return &otelcol.ConsumerArguments{
		Metrics: []otelcol.Consumer{&metricConsumer},
	}
}

// metricJson should match format from the protobuf definition:
// https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/metrics/v1/metrics.proto
func CreateTestMetrics(metricJson string) pmetric.Metrics {
	decoder := &pmetric.JSONUnmarshaler{}
	data, err := decoder.UnmarshalMetrics([]byte(metricJson))
	if err != nil {
		panic(err)
	}
	return data
}

func marshalMetrics(metrics pmetric.Metrics) string {
	marshaler := &pmetric.JSONMarshaler{}
	data, err := marshaler.MarshalMetrics(metrics)
	if err != nil {
		panic(err)
	}
	return string(data)
}
