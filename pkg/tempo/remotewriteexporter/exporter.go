package remotewriteexporter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/storage"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.uber.org/atomic"
)

const (
	nameLabelKey  = "__name__"
	sumSuffix     = "sum"
	countSuffix   = "count"
	bucketSuffix  = "bucket"
	leStr         = "le"
	infBucket     = "+Inf"
	counterSuffix = "total"
	noSuffix      = ""
)

type dataPoint interface {
	LabelsMap() pdata.StringMap
	Timestamp() pdata.TimestampUnixNano
}

type remoteWriteExporter struct {
	done atomic.Bool
	prom *instance.Instance

	constLabels labels.Labels
	namespace   string
}

func newRemoteWriteExporter(cfg *Config) (component.MetricsExporter, error) {
	return &remoteWriteExporter{
		done:        atomic.Bool{},
		constLabels: cfg.ConstLabels,
		namespace:   cfg.Namespace,
	}, nil
}

func (e *remoteWriteExporter) Start(ctx context.Context, _ component.Host) error {
	prom := ctx.Value(contextkeys.Prometheus).(*instance.Instance)
	if prom == nil {
		return fmt.Errorf("key does not contain a Prometheus instance")
	}
	e.prom = prom
	return nil
}

func (e *remoteWriteExporter) Shutdown(_ context.Context) error {
	e.done.Store(true)
	return nil
}

func (e *remoteWriteExporter) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	if e.done.Load() {
		return nil
	}

	app := e.prom.Appender(ctx)

	rm := md.ResourceMetrics()
	for i := 0; i < rm.Len(); i++ {
		ilm := rm.At(i).InstrumentationLibraryMetrics()
		for j := 0; j < ilm.Len(); j++ {
			ms := ilm.At(j).Metrics()
			for k := 0; k < ms.Len(); k++ {
				switch m := ms.At(k); m.DataType() {
				case pdata.MetricDataTypeDoubleSum, pdata.MetricDataTypeIntSum, pdata.MetricDataTypeDoubleGauge, pdata.MetricDataTypeIntGauge:
					if err := e.processScalarMetric(app, m); err != nil {
						return fmt.Errorf("failed to process metric %s", err)
					}
				case pdata.MetricDataTypeDoubleHistogram, pdata.MetricDataTypeIntHistogram:
					if err := e.processHistogramMetrics(app, m); err != nil {
						return fmt.Errorf("failed to process metric %s", err)
					}
				case pdata.MetricDataTypeDoubleSummary:
					return fmt.Errorf("%s processing unimplemented", m.DataType())
				default:
					return fmt.Errorf("unsupported m data type %s", m.DataType())
				}
			}
		}
	}

	return app.Commit()
}

func (e *remoteWriteExporter) processHistogramMetrics(app storage.Appender, m pdata.Metric) error {
	switch m.DataType() {
	case pdata.MetricDataTypeIntHistogram:
		dps := m.IntHistogram().DataPoints()
		if err := e.handleHistogramIntDataPoints(app, m.Name(), dps); err != nil {
			return nil
		}
	case pdata.MetricDataTypeDoubleHistogram:
		return fmt.Errorf("unsupported metric data type %s", m.DataType().String())
	}
	return nil
}

func (e *remoteWriteExporter) handleHistogramIntDataPoints(app storage.Appender, name string, dataPoints pdata.IntHistogramDataPointSlice) error {
	for ix := 0; ix < dataPoints.Len(); ix++ {
		dataPoint := dataPoints.At(ix)
		if err := e.appendDataPoint(app, name, sumSuffix, dataPoint, float64(dataPoint.Sum())); err != nil {
			return err
		}
		if err := e.appendDataPoint(app, name, countSuffix, dataPoint, float64(dataPoint.Count())); err != nil {
			return err
		}

		var cumulativeCount uint64
		for ix, eb := range dataPoint.ExplicitBounds() {
			if ix >= len(dataPoint.BucketCounts()) {
				break
			}
			cumulativeCount += dataPoint.BucketCounts()[ix]
			boundStr := strconv.FormatFloat(eb, 'f', -1, 64)
			ls := labels.Labels{{Name: leStr, Value: boundStr}}
			if err := e.appendDataPointWithLabels(app, name, bucketSuffix, dataPoint, float64(dataPoint.Count()), ls); err != nil {
				return err
			}
		}
		// add le=+Inf bucket
		cumulativeCount += dataPoint.BucketCounts()[len(dataPoint.BucketCounts())-1]
		ls := labels.Labels{{Name: leStr, Value: infBucket}}
		if err := e.appendDataPointWithLabels(app, name, bucketSuffix, dataPoint, float64(dataPoint.Count()), ls); err != nil {
			return err
		}

	}
	return nil
}

func (e *remoteWriteExporter) processScalarMetric(app storage.Appender, m pdata.Metric) error {
	switch m.DataType() {
	case pdata.MetricDataTypeIntSum:
		dataPoints := m.IntSum().DataPoints()
		if err := e.handleScalarIntDataPoints(app, m.Name(), counterSuffix, dataPoints); err != nil {
			return err
		}
	case pdata.MetricDataTypeDoubleSum:
		dataPoints := m.DoubleSum().DataPoints()
		if err := e.handleScalarFloatDataPoints(app, m.Name(), counterSuffix, dataPoints); err != nil {
			return err
		}
	case pdata.MetricDataTypeIntGauge:
		dataPoints := m.IntGauge().DataPoints()
		if err := e.handleScalarIntDataPoints(app, m.Name(), noSuffix, dataPoints); err != nil {
			return err
		}
	case pdata.MetricDataTypeDoubleGauge:
		dataPoints := m.DoubleGauge().DataPoints()
		if err := e.handleScalarFloatDataPoints(app, m.Name(), noSuffix, dataPoints); err != nil {
			return err
		}
	}
	return nil
}

func (e *remoteWriteExporter) handleScalarIntDataPoints(app storage.Appender, name, suffix string, dataPoints pdata.IntDataPointSlice) error {
	for ix := 0; ix < dataPoints.Len(); ix++ {
		dataPoint := dataPoints.At(ix)
		if err := e.appendDataPoint(app, name, suffix, dataPoint, float64(dataPoint.Value())); err != nil {
			return err
		}
	}
	return nil
}

func (e *remoteWriteExporter) handleScalarFloatDataPoints(app storage.Appender, name, suffix string, dataPoints pdata.DoubleDataPointSlice) error {
	for ix := 0; ix < dataPoints.Len(); ix++ {
		dataPoint := dataPoints.At(ix)
		if err := e.appendDataPoint(app, name, suffix, dataPoint, dataPoint.Value()); err != nil {
			return err
		}
	}
	return nil
}

func (e *remoteWriteExporter) appendDataPoint(app storage.Appender, name, suffix string, dp dataPoint, v float64) error {
	return e.appendDataPointWithLabels(app, name, suffix, dp, v, labels.Labels{})
}

func (e *remoteWriteExporter) appendDataPointWithLabels(app storage.Appender, name, suffix string, dp dataPoint, v float64, customLabels labels.Labels) error {
	ls := e.createLabelSet(name, suffix, dp.LabelsMap(), customLabels)
	ts := timestamp.FromTime(time.Unix(0, int64(dp.Timestamp())))
	if _, err := app.Append(0, ls, ts, v); err != nil {
		return err
	}
	return nil
}

func (e *remoteWriteExporter) createLabelSet(name, suffix string, labelMap pdata.StringMap, customLabels labels.Labels) labels.Labels {
	ls := make(labels.Labels, 0, labelMap.Len()+1+len(e.constLabels)+len(customLabels))
	// Labels from spanmetrics processor
	labelMap.ForEach(func(k string, v string) {
		ls = append(ls, labels.Label{
			Name:  strings.Replace(k, ".", "_", -1),
			Value: v,
		})
	})
	// Metric name label
	ls = append(ls, labels.Label{
		Name:  nameLabelKey,
		Value: metricName(e.namespace, name, suffix),
	})
	// Const labels
	ls = append(ls, e.constLabels...)
	// Custom labels
	ls = append(ls, customLabels...)
	return ls
}

func metricName(namespace, metric, suffix string) string {
	if len(suffix) != 0 {
		return fmt.Sprintf("%s_%s_%s", namespace, metric, suffix)
	}
	return fmt.Sprintf("%s_%s", namespace, metric)
}
