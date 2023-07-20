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
	"math"
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/scrape"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type testMetadataStore map[string]scrape.MetricMetadata

func (tmc testMetadataStore) GetMetadata(familyName string) (scrape.MetricMetadata, bool) {
	lookup, ok := tmc[familyName]
	return lookup, ok
}

func (tmc testMetadataStore) ListMetadata() []scrape.MetricMetadata { return nil }

func (tmc testMetadataStore) SizeMetadata() int { return 0 }

func (tmc testMetadataStore) LengthMetadata() int {
	return len(tmc)
}

var mc = testMetadataStore{
	"counter": scrape.MetricMetadata{
		Metric: "cr",
		Type:   textparse.MetricTypeCounter,
		Help:   "This is some help for a counter",
		Unit:   "By",
	},
	"gauge": scrape.MetricMetadata{
		Metric: "ge",
		Type:   textparse.MetricTypeGauge,
		Help:   "This is some help for a gauge",
		Unit:   "1",
	},
	"gaugehistogram": scrape.MetricMetadata{
		Metric: "gh",
		Type:   textparse.MetricTypeGaugeHistogram,
		Help:   "This is some help for a gauge histogram",
		Unit:   "?",
	},
	"histogram": scrape.MetricMetadata{
		Metric: "hg",
		Type:   textparse.MetricTypeHistogram,
		Help:   "This is some help for a histogram",
		Unit:   "ms",
	},
	"histogram_stale": scrape.MetricMetadata{
		Metric: "hg_stale",
		Type:   textparse.MetricTypeHistogram,
		Help:   "This is some help for a histogram",
		Unit:   "ms",
	},
	"summary": scrape.MetricMetadata{
		Metric: "s",
		Type:   textparse.MetricTypeSummary,
		Help:   "This is some help for a summary",
		Unit:   "ms",
	},
	"summary_stale": scrape.MetricMetadata{
		Metric: "s_stale",
		Type:   textparse.MetricTypeSummary,
		Help:   "This is some help for a summary",
		Unit:   "ms",
	},
	"unknown": scrape.MetricMetadata{
		Metric: "u",
		Type:   textparse.MetricTypeUnknown,
		Help:   "This is some help for an unknown metric",
		Unit:   "?",
	},
}

func TestMetricGroupData_toDistributionUnitTest(t *testing.T) {
	type scrape struct {
		at         int64
		value      float64
		metric     string
		extraLabel labels.Label
	}
	tests := []struct {
		name                string
		metricName          string
		labels              labels.Labels
		scrapes             []*scrape
		want                func() pmetric.HistogramDataPoint
		wantErr             bool
		intervalStartTimeMs int64
	}{
		{
			name:                "histogram with startTimestamp",
			metricName:          "histogram",
			intervalStartTimeMs: 11,
			labels:              labels.FromMap(map[string]string{"a": "A", "b": "B"}),
			scrapes: []*scrape{
				{at: 11, value: 66, metric: "histogram_count"},
				{at: 11, value: 1004.78, metric: "histogram_sum"},
				{at: 11, value: 33, metric: "histogram_bucket", extraLabel: labels.Label{Name: "le", Value: "0.75"}},
				{at: 11, value: 55, metric: "histogram_bucket", extraLabel: labels.Label{Name: "le", Value: "2.75"}},
				{at: 11, value: 66, metric: "histogram_bucket", extraLabel: labels.Label{Name: "le", Value: "+Inf"}},
			},
			want: func() pmetric.HistogramDataPoint {
				point := pmetric.NewHistogramDataPoint()
				point.SetCount(66)
				point.SetSum(1004.78)
				point.SetTimestamp(pcommon.Timestamp(11 * time.Millisecond)) // the time in milliseconds -> nanoseconds.
				point.ExplicitBounds().FromRaw([]float64{0.75, 2.75})
				point.BucketCounts().FromRaw([]uint64{33, 22, 11})
				point.SetStartTimestamp(pcommon.Timestamp(11 * time.Millisecond)) // the time in milliseconds -> nanoseconds.
				attributes := point.Attributes()
				attributes.PutStr("a", "A")
				attributes.PutStr("b", "B")
				return point
			},
		},
		{
			name:                "histogram that is stale",
			metricName:          "histogram_stale",
			intervalStartTimeMs: 11,
			labels:              labels.FromMap(map[string]string{"a": "A", "b": "B"}),
			scrapes: []*scrape{
				{at: 11, value: math.Float64frombits(value.StaleNaN), metric: "histogram_stale_count"},
				{at: 11, value: math.Float64frombits(value.StaleNaN), metric: "histogram_stale_sum"},
				{at: 11, value: math.Float64frombits(value.StaleNaN), metric: "histogram_bucket", extraLabel: labels.Label{Name: "le", Value: "0.75"}},
				{at: 11, value: math.Float64frombits(value.StaleNaN), metric: "histogram_bucket", extraLabel: labels.Label{Name: "le", Value: "2.75"}},
				{at: 11, value: math.Float64frombits(value.StaleNaN), metric: "histogram_bucket", extraLabel: labels.Label{Name: "le", Value: "+Inf"}},
			},
			want: func() pmetric.HistogramDataPoint {
				point := pmetric.NewHistogramDataPoint()
				point.SetTimestamp(pcommon.Timestamp(11 * time.Millisecond)) // the time in milliseconds -> nanoseconds.
				point.SetFlags(pmetric.DefaultDataPointFlags.WithNoRecordedValue(true))
				point.ExplicitBounds().FromRaw([]float64{0.75, 2.75})
				point.BucketCounts().FromRaw([]uint64{0, 0, 0})
				point.SetStartTimestamp(pcommon.Timestamp(11 * time.Millisecond)) // the time in milliseconds -> nanoseconds.
				attributes := point.Attributes()
				attributes.PutStr("a", "A")
				attributes.PutStr("b", "B")
				return point
			},
		},
		{
			name:                "histogram with inconsistent timestamps",
			metricName:          "histogram_inconsistent_ts",
			intervalStartTimeMs: 11,
			labels:              labels.FromMap(map[string]string{"a": "A", "le": "0.75", "b": "B"}),
			scrapes: []*scrape{
				{at: 11, value: math.Float64frombits(value.StaleNaN), metric: "histogram_stale_count"},
				{at: 12, value: math.Float64frombits(value.StaleNaN), metric: "histogram_stale_sum"},
				{at: 13, value: math.Float64frombits(value.StaleNaN), metric: "value"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mp := newMetricFamily(tt.metricName, mc, zap.NewNop())
			for i, tv := range tt.scrapes {
				var lbls labels.Labels
				if tv.extraLabel.Name != "" {
					lbls = labels.NewBuilder(tt.labels).Set(tv.extraLabel.Name, tv.extraLabel.Value).Labels()
				} else {
					lbls = tt.labels.Copy()
				}
				err := mp.Add(tv.metric, lbls, tv.at, tv.value)
				if tt.wantErr {
					if i != 0 {
						require.Error(t, err)
					}
				} else {
					require.NoError(t, err)
				}
			}
			if tt.wantErr {
				// Don't check the result if we got an error
				return
			}

			require.Len(t, mp.groups, 1)
			groupKey := mp.getGroupKey(tt.labels.Copy())
			require.NotNil(t, mp.groups[groupKey])

			sl := pmetric.NewMetricSlice()
			mp.appendMetric(sl)

			require.Equal(t, 1, sl.Len(), "Exactly one metric expected")
			metric := sl.At(0)
			require.Equal(t, mc[tt.metricName].Help, metric.Description(), "Expected help metadata in metric description")
			require.Equal(t, mc[tt.metricName].Unit, metric.Unit(), "Expected unit metadata in metric")

			hdpL := metric.Histogram().DataPoints()
			require.Equal(t, 1, hdpL.Len(), "Exactly one point expected")
			got := hdpL.At(0)
			want := tt.want()
			require.Equal(t, want, got, "Expected the points to be equal")
		})
	}
}

func TestMetricGroupData_toSummaryUnitTest(t *testing.T) {
	type scrape struct {
		at     int64
		value  float64
		metric string
	}

	type labelsScrapes struct {
		labels  labels.Labels
		scrapes []*scrape
	}
	tests := []struct {
		name          string
		labelsScrapes []*labelsScrapes
		want          func() pmetric.SummaryDataPoint
		wantErr       bool
	}{
		{
			name: "summary",
			labelsScrapes: []*labelsScrapes{
				{
					labels: labels.FromMap(map[string]string{"a": "A", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 10, metric: "summary_count"},
						{at: 14, value: 15, metric: "summary_sum"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.0", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 8, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.75", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 33.7, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.50", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 27, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.90", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 56, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.99", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 82, metric: "value"},
					},
				},
			},
			want: func() pmetric.SummaryDataPoint {
				point := pmetric.NewSummaryDataPoint()
				point.SetCount(10)
				point.SetSum(15)
				qtL := point.QuantileValues()
				qn0 := qtL.AppendEmpty()
				qn0.SetQuantile(0)
				qn0.SetValue(8)
				qn50 := qtL.AppendEmpty()
				qn50.SetQuantile(.5)
				qn50.SetValue(27)
				qn75 := qtL.AppendEmpty()
				qn75.SetQuantile(.75)
				qn75.SetValue(33.7)
				qn90 := qtL.AppendEmpty()
				qn90.SetQuantile(.9)
				qn90.SetValue(56)
				qn99 := qtL.AppendEmpty()
				qn99.SetQuantile(.99)
				qn99.SetValue(82)
				point.SetTimestamp(pcommon.Timestamp(14 * time.Millisecond))      // the time in milliseconds -> nanoseconds.
				point.SetStartTimestamp(pcommon.Timestamp(14 * time.Millisecond)) // the time in milliseconds -> nanoseconds
				attributes := point.Attributes()
				attributes.PutStr("a", "A")
				attributes.PutStr("b", "B")
				return point
			},
		},
		{
			name: "summary_stale",
			labelsScrapes: []*labelsScrapes{
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.0", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 10, metric: "summary_stale_count"},
						{at: 14, value: 12, metric: "summary_stale_sum"},
						{at: 14, value: 8, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.75", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 10, metric: "summary_stale_count"},
						{at: 14, value: 1004.78, metric: "summary_stale_sum"},
						{at: 14, value: 33.7, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.50", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 10, metric: "summary_stale_count"},
						{at: 14, value: 13, metric: "summary_stale_sum"},
						{at: 14, value: 27, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.90", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: 10, metric: "summary_stale_count"},
						{at: 14, value: 14, metric: "summary_stale_sum"},
						{at: 14, value: 56, metric: "value"},
					},
				},
				{
					labels: labels.FromMap(map[string]string{"a": "A", "quantile": "0.99", "b": "B"}),
					scrapes: []*scrape{
						{at: 14, value: math.Float64frombits(value.StaleNaN), metric: "summary_stale_count"},
						{at: 14, value: math.Float64frombits(value.StaleNaN), metric: "summary_stale_sum"},
						{at: 14, value: math.Float64frombits(value.StaleNaN), metric: "value"},
					},
				},
			},
			want: func() pmetric.SummaryDataPoint {
				point := pmetric.NewSummaryDataPoint()
				qtL := point.QuantileValues()
				qn0 := qtL.AppendEmpty()
				point.SetFlags(pmetric.DefaultDataPointFlags.WithNoRecordedValue(true))
				qn0.SetQuantile(0)
				qn0.SetValue(0)
				qn50 := qtL.AppendEmpty()
				qn50.SetQuantile(.5)
				qn50.SetValue(0)
				qn75 := qtL.AppendEmpty()
				qn75.SetQuantile(.75)
				qn75.SetValue(0)
				qn90 := qtL.AppendEmpty()
				qn90.SetQuantile(.9)
				qn90.SetValue(0)
				qn99 := qtL.AppendEmpty()
				qn99.SetQuantile(.99)
				qn99.SetValue(0)
				point.SetTimestamp(pcommon.Timestamp(14 * time.Millisecond))      // the time in milliseconds -> nanoseconds.
				point.SetStartTimestamp(pcommon.Timestamp(14 * time.Millisecond)) // the time in milliseconds -> nanoseconds
				attributes := point.Attributes()
				attributes.PutStr("a", "A")
				attributes.PutStr("b", "B")
				return point
			},
		},
		{
			name: "summary with inconsistent timestamps",
			labelsScrapes: []*labelsScrapes{
				{
					labels: labels.FromMap(map[string]string{"a": "A", "b": "B"}),
					scrapes: []*scrape{
						{at: 11, value: 10, metric: "summary_count"},
						{at: 14, value: 15, metric: "summary_sum"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mp := newMetricFamily(tt.name, mc, zap.NewNop())
			for _, lbs := range tt.labelsScrapes {
				for i, scrape := range lbs.scrapes {
					err := mp.Add(scrape.metric, lbs.labels.Copy(), scrape.at, scrape.value)
					if tt.wantErr {
						// The first scrape won't have an error
						if i != 0 {
							require.Error(t, err)
						}
					} else {
						require.NoError(t, err)
					}
				}
			}
			if tt.wantErr {
				// Don't check the result if we got an error
				return
			}

			require.Len(t, mp.groups, 1)
			groupKey := mp.getGroupKey(tt.labelsScrapes[0].labels.Copy())
			require.NotNil(t, mp.groups[groupKey])

			sl := pmetric.NewMetricSlice()
			mp.appendMetric(sl)

			require.Equal(t, 1, sl.Len(), "Exactly one metric expected")
			metric := sl.At(0)
			require.Equal(t, mc[tt.name].Help, metric.Description(), "Expected help metadata in metric description")
			require.Equal(t, mc[tt.name].Unit, metric.Unit(), "Expected unit metadata in metric")

			sdpL := metric.Summary().DataPoints()
			require.Equal(t, 1, sdpL.Len(), "Exactly one point expected")
			got := sdpL.At(0)
			want := tt.want()
			require.Equal(t, want, got, "Expected the points to be equal")
		})
	}
}

func TestMetricGroupData_toNumberDataUnitTest(t *testing.T) {
	type scrape struct {
		at     int64
		value  float64
		metric string
	}
	tests := []struct {
		name                     string
		metricKind               string
		labels                   labels.Labels
		scrapes                  []*scrape
		intervalStartTimestampMs int64
		want                     func() pmetric.NumberDataPoint
	}{
		{
			metricKind:               "counter",
			name:                     "counter:: startTimestampMs of 11",
			intervalStartTimestampMs: 11,
			labels:                   labels.FromMap(map[string]string{"a": "A", "b": "B"}),
			scrapes: []*scrape{
				{at: 13, value: 33.7, metric: "value"},
			},
			want: func() pmetric.NumberDataPoint {
				point := pmetric.NewNumberDataPoint()
				point.SetDoubleValue(33.7)
				point.SetTimestamp(pcommon.Timestamp(13 * time.Millisecond))      // the time in milliseconds -> nanoseconds.
				point.SetStartTimestamp(pcommon.Timestamp(13 * time.Millisecond)) // the time in milliseconds -> nanoseconds.
				attributes := point.Attributes()
				attributes.PutStr("a", "A")
				attributes.PutStr("b", "B")
				return point
			},
		},
		{
			name:                     "counter:: startTimestampMs of 0",
			metricKind:               "counter",
			intervalStartTimestampMs: 0,
			labels:                   labels.FromMap(map[string]string{"a": "A", "b": "B"}),
			scrapes: []*scrape{
				{at: 28, value: 99.9, metric: "value"},
			},
			want: func() pmetric.NumberDataPoint {
				point := pmetric.NewNumberDataPoint()
				point.SetDoubleValue(99.9)
				point.SetTimestamp(pcommon.Timestamp(28 * time.Millisecond))      // the time in milliseconds -> nanoseconds.
				point.SetStartTimestamp(pcommon.Timestamp(28 * time.Millisecond)) // the time in milliseconds -> nanoseconds.
				attributes := point.Attributes()
				attributes.PutStr("a", "A")
				attributes.PutStr("b", "B")
				return point
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mp := newMetricFamily(tt.metricKind, mc, zap.NewNop())
			for _, tv := range tt.scrapes {
				require.NoError(t, mp.Add(tv.metric, tt.labels.Copy(), tv.at, tv.value))
			}

			require.Len(t, mp.groups, 1)
			groupKey := mp.getGroupKey(tt.labels.Copy())
			require.NotNil(t, mp.groups[groupKey])

			sl := pmetric.NewMetricSlice()
			mp.appendMetric(sl)

			require.Equal(t, 1, sl.Len(), "Exactly one metric expected")
			metric := sl.At(0)
			require.Equal(t, mc[tt.metricKind].Help, metric.Description(), "Expected help metadata in metric description")
			require.Equal(t, mc[tt.metricKind].Unit, metric.Unit(), "Expected unit metadata in metric")

			ndpL := metric.Sum().DataPoints()
			require.Equal(t, 1, ndpL.Len(), "Exactly one point expected")
			got := ndpL.At(0)
			want := tt.want()
			require.Equal(t, want, got, "Expected the points to be equal")
		})
	}
}
