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

package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/scrape"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

const (
	startTimestamp      = pcommon.Timestamp(1555366608340000000)
	ts                  = int64(1555366610000)
	interval            = int64(15 * 1000)
	tsNanos             = pcommon.Timestamp(ts * 1e6)
	tsPlusIntervalNanos = pcommon.Timestamp((ts + interval) * 1e6)
)

var (
	target = scrape.NewTarget(
		// processedLabels contain label values after processing (e.g. relabeling)
		labels.FromMap(map[string]string{
			model.InstanceLabel: "localhost:8080",
		}),
		// discoveredLabels contain labels prior to any processing
		labels.FromMap(map[string]string{
			model.AddressLabel: "address:8080",
			model.SchemeLabel:  "http",
		}),
		nil)

	scrapeCtx = scrape.ContextWithMetricMetadataStore(
		scrape.ContextWithTarget(context.Background(), target),
		testMetadataStore(testMetadata))
)

func TestTransactionCommitWithoutAdding(t *testing.T) {
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, consumertest.NewNop(), nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	assert.NoError(t, tr.Commit())
}

func TestTransactionRollbackDoesNothing(t *testing.T) {
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, consumertest.NewNop(), nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	assert.NoError(t, tr.Rollback())
}

func TestTransactionUpdateMetadataDoesNothing(t *testing.T) {
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, consumertest.NewNop(), nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	_, err := tr.UpdateMetadata(0, labels.New(), metadata.Metadata{})
	assert.NoError(t, err)
}

func TestTransactionAppendNoTarget(t *testing.T) {
	badLabels := labels.FromStrings(model.MetricNameLabel, "counter_test")
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, consumertest.NewNop(), nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	_, err := tr.Append(0, badLabels, time.Now().Unix()*1000, 1.0)
	assert.Error(t, err)
}

func TestTransactionAppendNoMetricName(t *testing.T) {
	jobNotFoundLb := labels.FromMap(map[string]string{
		model.InstanceLabel: "localhost:8080",
		model.JobLabel:      "test2",
	})
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, consumertest.NewNop(), nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	_, err := tr.Append(0, jobNotFoundLb, time.Now().Unix()*1000, 1.0)
	assert.ErrorIs(t, err, errMetricNameNotFound)

	assert.ErrorIs(t, tr.Commit(), errNoDataToBuild)
}

func TestTransactionAppendEmptyMetricName(t *testing.T) {
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, consumertest.NewNop(), nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	_, err := tr.Append(0, labels.FromMap(map[string]string{
		model.InstanceLabel:   "localhost:8080",
		model.JobLabel:        "test2",
		model.MetricNameLabel: "",
	}), time.Now().Unix()*1000, 1.0)
	assert.ErrorIs(t, err, errMetricNameNotFound)
}

func TestTransactionAppendResource(t *testing.T) {
	sink := new(consumertest.MetricsSink)
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, sink, nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	_, err := tr.Append(0, labels.FromMap(map[string]string{
		model.InstanceLabel:   "localhost:8080",
		model.JobLabel:        "test",
		model.MetricNameLabel: "counter_test",
	}), time.Now().Unix()*1000, 1.0)
	assert.NoError(t, err)
	_, err = tr.Append(0, labels.FromMap(map[string]string{
		model.InstanceLabel:   "localhost:8080",
		model.JobLabel:        "test",
		model.MetricNameLabel: startTimeMetricName,
	}), time.Now().UnixMilli(), 1.0)
	assert.NoError(t, err)
	assert.NoError(t, tr.Commit())
	expectedResource := CreateResource("test", "localhost:8080", labels.FromStrings(model.SchemeLabel, "http"))
	mds := sink.AllMetrics()
	require.Len(t, mds, 1)
	gotResource := mds[0].ResourceMetrics().At(0).Resource()
	require.Equal(t, expectedResource, gotResource)
}

func TestTransactionCommitErrorWhenAdjusterError(t *testing.T) {
	goodLabels := labels.FromMap(map[string]string{
		model.InstanceLabel:   "localhost:8080",
		model.JobLabel:        "test",
		model.MetricNameLabel: "counter_test",
	})
	sink := new(consumertest.MetricsSink)
	adjusterErr := errors.New("adjuster error")
	tr := newTransaction(scrapeCtx, &errorAdjuster{err: adjusterErr}, sink, nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
	_, err := tr.Append(0, goodLabels, time.Now().Unix()*1000, 1.0)
	assert.NoError(t, err)
	assert.ErrorIs(t, tr.Commit(), adjusterErr)
}

// Ensure that we reject duplicate label keys. See https://github.com/open-telemetry/wg-prometheus/issues/44.
func TestTransactionAppendDuplicateLabels(t *testing.T) {
	sink := new(consumertest.MetricsSink)
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, sink, nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))

	dupLabels := labels.FromStrings(
		model.InstanceLabel, "0.0.0.0:8855",
		model.JobLabel, "test",
		model.MetricNameLabel, "counter_test",
		"a", "1",
		"a", "6",
		"z", "9",
	)

	_, err := tr.Append(0, dupLabels, 1917, 1.0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid sample: non-unique label names: "a"`)
}

func TestTransactionAppendHistogramNoLe(t *testing.T) {
	sink := new(consumertest.MetricsSink)
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, sink, nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))

	goodLabels := labels.FromStrings(
		model.InstanceLabel, "0.0.0.0:8855",
		model.JobLabel, "test",
		model.MetricNameLabel, "hist_test_bucket",
	)

	_, err := tr.Append(0, goodLabels, 1917, 1.0)
	require.ErrorIs(t, err, errEmptyLeLabel)
}

func TestTransactionAppendSummaryNoQuantile(t *testing.T) {
	sink := new(consumertest.MetricsSink)
	tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, sink, nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))

	goodLabels := labels.FromStrings(
		model.InstanceLabel, "0.0.0.0:8855",
		model.JobLabel, "test",
		model.MetricNameLabel, "summary_test",
	)

	_, err := tr.Append(0, goodLabels, 1917, 1.0)
	require.ErrorIs(t, err, errEmptyQuantileLabel)
}

func nopObsRecv(t *testing.T) *obsreport.Receiver {
	res, err := obsreport.NewReceiver(obsreport.ReceiverSettings{
		ReceiverID:             component.NewID("prometheus"),
		Transport:              transport,
		ReceiverCreateSettings: receivertest.NewNopCreateSettings(),
	})

	assert.NoError(t, err)
	return res
}

func TestMetricBuilderCounters(t *testing.T) {
	tests := []buildTestData{
		{
			name: "single-item",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("counter_test", 100, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("counter_test")
				sum := m0.SetEmptySum()
				sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				sum.SetIsMonotonic(true)
				pt0 := sum.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "two-items",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("counter_test", 150, "foo", "bar"),
						createDataPoint("counter_test", 25, "foo", "other"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("counter_test")
				sum := m0.SetEmptySum()
				sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				sum.SetIsMonotonic(true)
				pt0 := sum.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(150.0)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				pt1 := sum.DataPoints().AppendEmpty()
				pt1.SetDoubleValue(25.0)
				pt1.SetStartTimestamp(startTimestamp)
				pt1.SetTimestamp(tsNanos)
				pt1.Attributes().PutStr("foo", "other")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "two-metrics",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("counter_test", 150, "foo", "bar"),
						createDataPoint("counter_test", 25, "foo", "other"),
						createDataPoint("counter_test2", 100, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("counter_test")
				sum0 := m0.SetEmptySum()
				sum0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				sum0.SetIsMonotonic(true)
				pt0 := sum0.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(150.0)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				pt1 := sum0.DataPoints().AppendEmpty()
				pt1.SetDoubleValue(25.0)
				pt1.SetStartTimestamp(startTimestamp)
				pt1.SetTimestamp(tsNanos)
				pt1.Attributes().PutStr("foo", "other")

				m1 := mL0.AppendEmpty()
				m1.SetName("counter_test2")
				sum1 := m1.SetEmptySum()
				sum1.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				sum1.SetIsMonotonic(true)
				pt2 := sum1.DataPoints().AppendEmpty()
				pt2.SetDoubleValue(100.0)
				pt2.SetStartTimestamp(startTimestamp)
				pt2.SetTimestamp(tsNanos)
				pt2.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "metrics-with-poor-names",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("poor_name_count", 100, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("poor_name_count")
				sum := m0.SetEmptySum()
				sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				sum.SetIsMonotonic(true)
				pt0 := sum.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestMetricBuilderGauges(t *testing.T) {
	tests := []buildTestData{
		{
			name: "one-gauge",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("gauge_test", 100, "foo", "bar"),
					},
				},
				{
					pts: []*testDataPoint{
						createDataPoint("gauge_test", 90, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("gauge_test")
				gauge0 := m0.SetEmptyGauge()
				pt0 := gauge0.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetStartTimestamp(0)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				md1 := pmetric.NewMetrics()
				mL1 := md1.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m1 := mL1.AppendEmpty()
				m1.SetName("gauge_test")
				gauge1 := m1.SetEmptyGauge()
				pt1 := gauge1.DataPoints().AppendEmpty()
				pt1.SetDoubleValue(90.0)
				pt1.SetStartTimestamp(0)
				pt1.SetTimestamp(tsPlusIntervalNanos)
				pt1.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0, md1}
			},
		},
		{
			name: "gauge-with-different-tags",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("gauge_test", 100, "foo", "bar"),
						createDataPoint("gauge_test", 200, "bar", "foo"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("gauge_test")
				gauge0 := m0.SetEmptyGauge()
				pt0 := gauge0.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetStartTimestamp(0)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				pt1 := gauge0.DataPoints().AppendEmpty()
				pt1.SetDoubleValue(200.0)
				pt1.SetStartTimestamp(0)
				pt1.SetTimestamp(tsNanos)
				pt1.Attributes().PutStr("bar", "foo")

				return []pmetric.Metrics{md0}
			},
		},
		{
			// TODO: A decision need to be made. If we want to have the behavior which can generate different tag key
			//  sets because metrics come and go
			name: "gauge-comes-and-go-with-different-tagset",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("gauge_test", 100, "foo", "bar"),
						createDataPoint("gauge_test", 200, "bar", "foo"),
					},
				},
				{
					pts: []*testDataPoint{
						createDataPoint("gauge_test", 20, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("gauge_test")
				gauge0 := m0.SetEmptyGauge()
				pt0 := gauge0.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetStartTimestamp(0)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				pt1 := gauge0.DataPoints().AppendEmpty()
				pt1.SetDoubleValue(200.0)
				pt1.SetStartTimestamp(0)
				pt1.SetTimestamp(tsNanos)
				pt1.Attributes().PutStr("bar", "foo")

				md1 := pmetric.NewMetrics()
				mL1 := md1.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m1 := mL1.AppendEmpty()
				m1.SetName("gauge_test")
				gauge1 := m1.SetEmptyGauge()
				pt2 := gauge1.DataPoints().AppendEmpty()
				pt2.SetDoubleValue(20.0)
				pt2.SetStartTimestamp(0)
				pt2.SetTimestamp(tsPlusIntervalNanos)
				pt2.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0, md1}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestMetricBuilderUntyped(t *testing.T) {
	tests := []buildTestData{
		{
			name: "one-unknown",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("unknown_test", 100, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("unknown_test")
				gauge0 := m0.SetEmptyGauge()
				pt0 := gauge0.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetStartTimestamp(0)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "no-type-hint",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("something_not_exists", 100, "foo", "bar"),
						createDataPoint("theother_not_exists", 200, "foo", "bar"),
						createDataPoint("theother_not_exists", 300, "bar", "foo"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("something_not_exists")
				gauge0 := m0.SetEmptyGauge()
				pt0 := gauge0.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				m1 := mL0.AppendEmpty()
				m1.SetName("theother_not_exists")
				gauge1 := m1.SetEmptyGauge()
				pt1 := gauge1.DataPoints().AppendEmpty()
				pt1.SetDoubleValue(200.0)
				pt1.SetTimestamp(tsNanos)
				pt1.Attributes().PutStr("foo", "bar")

				pt2 := gauge1.DataPoints().AppendEmpty()
				pt2.SetDoubleValue(300.0)
				pt2.SetTimestamp(tsNanos)
				pt2.Attributes().PutStr("bar", "foo")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "untype-metric-poor-names",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("some_count", 100, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("some_count")
				gauge0 := m0.SetEmptyGauge()
				pt0 := gauge0.DataPoints().AppendEmpty()
				pt0.SetDoubleValue(100.0)
				pt0.SetTimestamp(tsNanos)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestMetricBuilderHistogram(t *testing.T) {
	tests := []buildTestData{
		{
			name: "single item",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 1, "foo", "bar", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "foo", "bar", "le", "20"),
						createDataPoint("hist_test_bucket", 10, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test_sum", 99, "foo", "bar"),
						createDataPoint("hist_test_count", 10, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("hist_test")
				hist0 := m0.SetEmptyHistogram()
				hist0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt0 := hist0.DataPoints().AppendEmpty()
				pt0.SetCount(10)
				pt0.SetSum(99)
				pt0.ExplicitBounds().FromRaw([]float64{10, 20})
				pt0.BucketCounts().FromRaw([]uint64{1, 1, 8})
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "multi-groups",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 1, "foo", "bar", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "foo", "bar", "le", "20"),
						createDataPoint("hist_test_bucket", 10, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test_sum", 99, "foo", "bar"),
						createDataPoint("hist_test_count", 10, "foo", "bar"),
						createDataPoint("hist_test_bucket", 1, "key2", "v2", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "key2", "v2", "le", "20"),
						createDataPoint("hist_test_bucket", 3, "key2", "v2", "le", "+inf"),
						createDataPoint("hist_test_sum", 50, "key2", "v2"),
						createDataPoint("hist_test_count", 3, "key2", "v2"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("hist_test")
				hist0 := m0.SetEmptyHistogram()
				hist0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt0 := hist0.DataPoints().AppendEmpty()
				pt0.SetCount(10)
				pt0.SetSum(99)
				pt0.ExplicitBounds().FromRaw([]float64{10, 20})
				pt0.BucketCounts().FromRaw([]uint64{1, 1, 8})
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.Attributes().PutStr("foo", "bar")

				pt1 := hist0.DataPoints().AppendEmpty()
				pt1.SetCount(3)
				pt1.SetSum(50)
				pt1.ExplicitBounds().FromRaw([]float64{10, 20})
				pt1.BucketCounts().FromRaw([]uint64{1, 1, 1})
				pt1.SetTimestamp(tsNanos)
				pt1.SetStartTimestamp(startTimestamp)
				pt1.Attributes().PutStr("key2", "v2")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "multi-groups-and-families",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 1, "foo", "bar", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "foo", "bar", "le", "20"),
						createDataPoint("hist_test_bucket", 10, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test_sum", 99, "foo", "bar"),
						createDataPoint("hist_test_count", 10, "foo", "bar"),
						createDataPoint("hist_test_bucket", 1, "key2", "v2", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "key2", "v2", "le", "20"),
						createDataPoint("hist_test_bucket", 3, "key2", "v2", "le", "+inf"),
						createDataPoint("hist_test_sum", 50, "key2", "v2"),
						createDataPoint("hist_test_count", 3, "key2", "v2"),
						createDataPoint("hist_test2_bucket", 1, "foo", "bar", "le", "10"),
						createDataPoint("hist_test2_bucket", 2, "foo", "bar", "le", "20"),
						createDataPoint("hist_test2_bucket", 3, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test2_sum", 50, "foo", "bar"),
						createDataPoint("hist_test2_count", 3, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("hist_test")
				hist0 := m0.SetEmptyHistogram()
				hist0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt0 := hist0.DataPoints().AppendEmpty()
				pt0.SetCount(10)
				pt0.SetSum(99)
				pt0.ExplicitBounds().FromRaw([]float64{10, 20})
				pt0.BucketCounts().FromRaw([]uint64{1, 1, 8})
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.Attributes().PutStr("foo", "bar")

				pt1 := hist0.DataPoints().AppendEmpty()
				pt1.SetCount(3)
				pt1.SetSum(50)
				pt1.ExplicitBounds().FromRaw([]float64{10, 20})
				pt1.BucketCounts().FromRaw([]uint64{1, 1, 1})
				pt1.SetTimestamp(tsNanos)
				pt1.SetStartTimestamp(startTimestamp)
				pt1.Attributes().PutStr("key2", "v2")

				m1 := mL0.AppendEmpty()
				m1.SetName("hist_test2")
				hist1 := m1.SetEmptyHistogram()
				hist1.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt2 := hist1.DataPoints().AppendEmpty()
				pt2.SetCount(3)
				pt2.SetSum(50)
				pt2.ExplicitBounds().FromRaw([]float64{10, 20})
				pt2.BucketCounts().FromRaw([]uint64{1, 1, 1})
				pt2.SetTimestamp(tsNanos)
				pt2.SetStartTimestamp(startTimestamp)
				pt2.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "unordered-buckets",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 10, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test_bucket", 1, "foo", "bar", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "foo", "bar", "le", "20"),
						createDataPoint("hist_test_sum", 99, "foo", "bar"),
						createDataPoint("hist_test_count", 10, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("hist_test")
				hist0 := m0.SetEmptyHistogram()
				hist0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt0 := hist0.DataPoints().AppendEmpty()
				pt0.SetCount(10)
				pt0.SetSum(99)
				pt0.ExplicitBounds().FromRaw([]float64{10, 20})
				pt0.BucketCounts().FromRaw([]uint64{1, 1, 8})
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			// this won't likely happen in real env, as prometheus won't generate histogram with less than 3 buckets
			name: "only-one-bucket",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 3, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test_count", 3, "foo", "bar"),
						createDataPoint("hist_test_sum", 100, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("hist_test")
				hist0 := m0.SetEmptyHistogram()
				hist0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt0 := hist0.DataPoints().AppendEmpty()
				pt0.SetCount(3)
				pt0.SetSum(100)
				pt0.BucketCounts().FromRaw([]uint64{3})
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			// this won't likely happen in real env, as prometheus won't generate histogram with less than 3 buckets
			name: "only-one-bucket-noninf",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 3, "foo", "bar", "le", "20"),
						createDataPoint("hist_test_count", 3, "foo", "bar"),
						createDataPoint("hist_test_sum", 100, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("hist_test")
				hist0 := m0.SetEmptyHistogram()
				hist0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt0 := hist0.DataPoints().AppendEmpty()
				pt0.SetCount(3)
				pt0.SetSum(100)
				pt0.BucketCounts().FromRaw([]uint64{3})
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "no-sum",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 1, "foo", "bar", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "foo", "bar", "le", "20"),
						createDataPoint("hist_test_bucket", 3, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test_count", 3, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("hist_test")
				hist0 := m0.SetEmptyHistogram()
				hist0.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
				pt0 := hist0.DataPoints().AppendEmpty()
				pt0.SetCount(3)
				pt0.ExplicitBounds().FromRaw([]float64{10, 20})
				pt0.BucketCounts().FromRaw([]uint64{1, 1, 1})
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "corrupted-no-buckets",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_sum", 99),
						createDataPoint("hist_test_count", 10),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				return []pmetric.Metrics{pmetric.NewMetrics()}
			},
		},
		{
			name: "corrupted-no-count",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("hist_test_bucket", 1, "foo", "bar", "le", "10"),
						createDataPoint("hist_test_bucket", 2, "foo", "bar", "le", "20"),
						createDataPoint("hist_test_bucket", 3, "foo", "bar", "le", "+inf"),
						createDataPoint("hist_test_sum", 99, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				return []pmetric.Metrics{pmetric.NewMetrics()}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestMetricBuilderSummary(t *testing.T) {
	tests := []buildTestData{
		{
			name: "no-sum-and-count",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("summary_test", 5, "foo", "bar", "quantile", "1"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				return []pmetric.Metrics{pmetric.NewMetrics()}
			},
		},
		{
			name: "no-count",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("summary_test", 1, "foo", "bar", "quantile", "0.5"),
						createDataPoint("summary_test", 2, "foo", "bar", "quantile", "0.75"),
						createDataPoint("summary_test", 5, "foo", "bar", "quantile", "1"),
						createDataPoint("summary_test_sum", 500, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				return []pmetric.Metrics{pmetric.NewMetrics()}
			},
		},
		{
			name: "no-sum",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("summary_test", 1, "foo", "bar", "quantile", "0.5"),
						createDataPoint("summary_test", 2, "foo", "bar", "quantile", "0.75"),
						createDataPoint("summary_test", 5, "foo", "bar", "quantile", "1"),
						createDataPoint("summary_test_count", 500, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("summary_test")
				sum0 := m0.SetEmptySummary()
				pt0 := sum0.DataPoints().AppendEmpty()
				pt0.SetTimestamp(tsNanos)
				pt0.SetStartTimestamp(startTimestamp)
				pt0.SetCount(500)
				pt0.SetSum(0.0)
				pt0.Attributes().PutStr("foo", "bar")
				qvL := pt0.QuantileValues()
				q50 := qvL.AppendEmpty()
				q50.SetQuantile(.50)
				q50.SetValue(1.0)
				q75 := qvL.AppendEmpty()
				q75.SetQuantile(.75)
				q75.SetValue(2.0)
				q100 := qvL.AppendEmpty()
				q100.SetQuantile(1)
				q100.SetValue(5.0)

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "empty-quantiles",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("summary_test_sum", 100, "foo", "bar"),
						createDataPoint("summary_test_count", 500, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("summary_test")
				sum0 := m0.SetEmptySummary()
				pt0 := sum0.DataPoints().AppendEmpty()
				pt0.SetStartTimestamp(startTimestamp)
				pt0.SetTimestamp(tsNanos)
				pt0.SetCount(500)
				pt0.SetSum(100.0)
				pt0.Attributes().PutStr("foo", "bar")

				return []pmetric.Metrics{md0}
			},
		},
		{
			name: "regular-summary",
			inputs: []*testScrapedPage{
				{
					pts: []*testDataPoint{
						createDataPoint("summary_test", 1, "foo", "bar", "quantile", "0.5"),
						createDataPoint("summary_test", 2, "foo", "bar", "quantile", "0.75"),
						createDataPoint("summary_test", 5, "foo", "bar", "quantile", "1"),
						createDataPoint("summary_test_sum", 100, "foo", "bar"),
						createDataPoint("summary_test_count", 500, "foo", "bar"),
					},
				},
			},
			wants: func() []pmetric.Metrics {
				md0 := pmetric.NewMetrics()
				mL0 := md0.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics()
				m0 := mL0.AppendEmpty()
				m0.SetName("summary_test")
				sum0 := m0.SetEmptySummary()
				pt0 := sum0.DataPoints().AppendEmpty()
				pt0.SetStartTimestamp(startTimestamp)
				pt0.SetTimestamp(tsNanos)
				pt0.SetCount(500)
				pt0.SetSum(100.0)
				pt0.Attributes().PutStr("foo", "bar")
				qvL := pt0.QuantileValues()
				q50 := qvL.AppendEmpty()
				q50.SetQuantile(.50)
				q50.SetValue(1.0)
				q75 := qvL.AppendEmpty()
				q75.SetQuantile(.75)
				q75.SetValue(2.0)
				q100 := qvL.AppendEmpty()
				q100.SetQuantile(1)
				q100.SetValue(5.0)

				return []pmetric.Metrics{md0}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

type buildTestData struct {
	name   string
	inputs []*testScrapedPage
	wants  func() []pmetric.Metrics
}

func (tt buildTestData) run(t *testing.T) {
	wants := tt.wants()
	assert.EqualValues(t, len(wants), len(tt.inputs))
	st := ts
	for i, page := range tt.inputs {
		sink := new(consumertest.MetricsSink)
		tr := newTransaction(scrapeCtx, &startTimeAdjuster{startTime: startTimestamp}, sink, nil, receivertest.NewNopCreateSettings(), nopObsRecv(t))
		for _, pt := range page.pts {
			// set ts for testing
			pt.t = st
			_, err := tr.Append(0, pt.lb, pt.t, pt.v)
			assert.NoError(t, err)
		}
		assert.NoError(t, tr.Commit())
		mds := sink.AllMetrics()
		if wants[i].ResourceMetrics().Len() == 0 {
			// Receiver does not emit empty metrics, so will not have anything in the sink.
			require.Len(t, mds, 0)
			st += interval
			continue
		}
		require.Len(t, mds, 1)
		assertEquivalentMetrics(t, wants[i], mds[0])
		st += interval
	}
}

type errorAdjuster struct {
	err error
}

func (ea *errorAdjuster) AdjustMetrics(pmetric.Metrics) error {
	return ea.err
}

type startTimeAdjuster struct {
	startTime pcommon.Timestamp
}

func (s *startTimeAdjuster) AdjustMetrics(metrics pmetric.Metrics) error {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			ilm := rm.ScopeMetrics().At(j)
			for k := 0; k < ilm.Metrics().Len(); k++ {
				metric := ilm.Metrics().At(k)
				switch metric.Type() {
				case pmetric.MetricTypeSum:
					dps := metric.Sum().DataPoints()
					for l := 0; l < dps.Len(); l++ {
						dps.At(l).SetStartTimestamp(s.startTime)
					}
				case pmetric.MetricTypeSummary:
					dps := metric.Summary().DataPoints()
					for l := 0; l < dps.Len(); l++ {
						dps.At(l).SetStartTimestamp(s.startTime)
					}
				case pmetric.MetricTypeHistogram:
					dps := metric.Histogram().DataPoints()
					for l := 0; l < dps.Len(); l++ {
						dps.At(l).SetStartTimestamp(s.startTime)
					}
				}
			}
		}
	}
	return nil
}

type testDataPoint struct {
	lb labels.Labels
	t  int64
	v  float64
}

type testScrapedPage struct {
	pts []*testDataPoint
}

func createDataPoint(mname string, value float64, tagPairs ...string) *testDataPoint {
	var lbls []string
	lbls = append(lbls, tagPairs...)
	lbls = append(lbls, model.MetricNameLabel, mname)
	lbls = append(lbls, model.JobLabel, "job")
	lbls = append(lbls, model.InstanceLabel, "instance")

	return &testDataPoint{
		lb: labels.FromStrings(lbls...),
		t:  ts,
		v:  value,
	}
}

func assertEquivalentMetrics(t *testing.T, want, got pmetric.Metrics) {
	require.Equal(t, want.ResourceMetrics().Len(), got.ResourceMetrics().Len())
	if want.ResourceMetrics().Len() == 0 {
		return
	}
	for i := 0; i < want.ResourceMetrics().Len(); i++ {
		wantSm := want.ResourceMetrics().At(i).ScopeMetrics()
		gotSm := got.ResourceMetrics().At(i).ScopeMetrics()
		require.Equal(t, wantSm.Len(), gotSm.Len())
		if wantSm.Len() == 0 {
			return
		}

		for j := 0; j < wantSm.Len(); j++ {
			wantMs := wantSm.At(j).Metrics()
			gotMs := gotSm.At(j).Metrics()
			require.Equal(t, wantMs.Len(), gotMs.Len())

			wmap := map[string]pmetric.Metric{}
			gmap := map[string]pmetric.Metric{}

			for k := 0; k < wantMs.Len(); k++ {
				wi := wantMs.At(k)
				wmap[wi.Name()] = wi
				gi := gotMs.At(k)
				gmap[gi.Name()] = gi
			}
			assert.EqualValues(t, wmap, gmap)
		}
	}
}
