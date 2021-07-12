package remotewriteexporter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/agent/pkg/prom/instance"
	"github.com/grafana/agent/pkg/tempo/contextkeys"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/storage"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
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
}

type remoteWriteExporter struct {
	done         atomic.Bool
	manager      instance.Manager
	promInstance string

	constLabels labels.Labels
	namespace   string

	logger log.Logger
}

func newRemoteWriteExporter(cfg *Config) (component.MetricsExporter, error) {
	logger := log.With(util.Logger, "component", "tempo remote write exporter")

	return &remoteWriteExporter{
		done:         atomic.Bool{},
		constLabels:  cfg.ConstLabels,
		namespace:    cfg.Namespace,
		promInstance: cfg.PromInstance,
		logger:       logger,
	}, nil
}

func (e *remoteWriteExporter) Start(ctx context.Context, _ component.Host) error {
	manager, ok := ctx.Value(contextkeys.Prometheus).(instance.Manager)
	if !ok || manager == nil {
		return fmt.Errorf("key does not contain a Prometheus instance")
	}
	e.manager = manager
	return nil
}

func (e *remoteWriteExporter) Shutdown(_ context.Context) error {
	e.done.Store(true)
	return nil
}

func (e *remoteWriteExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func (e *remoteWriteExporter) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	if e.done.Load() {
		return nil
	}

	prom, err := e.manager.GetInstance(e.promInstance)
	if err != nil {
		level.Warn(e.logger).Log("msg", "failed to get prom instance", "err", err)
		return nil
	}
	app := prom.Appender(ctx)

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
				case pdata.MetricDataTypeHistogram, pdata.MetricDataTypeIntHistogram:
					if err := e.processHistogramMetrics(app, m); err != nil {
						return fmt.Errorf("failed to process metric %s", err)
					}
				case pdata.MetricDataTypeSummary:
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
	case pdata.MetricDataTypeHistogram:
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
			if err := e.appendDataPointWithLabels(app, name, bucketSuffix, dataPoint, float64(cumulativeCount), ls); err != nil {
				return err
			}
		}
		// add le=+Inf bucket
		cumulativeCount += dataPoint.BucketCounts()[len(dataPoint.BucketCounts())-1]
		ls := labels.Labels{{Name: leStr, Value: infBucket}}
		if err := e.appendDataPointWithLabels(app, name, bucketSuffix, dataPoint, float64(cumulativeCount), ls); err != nil {
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
	// TODO(mario.rodriguez): Use timestamp from metric
	// time.Now() is used to avoid out-of-order metrics
	ts := timestamp.FromTime(time.Now())
	if _, err := app.Append(0, ls, ts, v); err != nil {
		return err
	}
	return nil
}

func (e *remoteWriteExporter) createLabelSet(name, suffix string, labelMap pdata.StringMap, customLabels labels.Labels) labels.Labels {
	ls := make(labels.Labels, 0, labelMap.Len()+1+len(e.constLabels)+len(customLabels))
	// Labels from spanmetrics processor
	labelMap.Range(func(k string, v string) bool {
		ls = append(ls, labels.Label{
			Name:  strings.Replace(k, ".", "_", -1),
			Value: v,
		})
		return true
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
