package remotewriteexporter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/metrics/instance"
	"github.com/grafana/agent/pkg/traces/contextkeys"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/atomic"
)

const (
	nameLabelKey = "__name__"
	sumSuffix    = "sum"
	countSuffix  = "count"
	bucketSuffix = "bucket"
	leStr        = "le"
	infBucket    = "+Inf"
	noSuffix     = ""
)

type remoteWriteExporter struct {
	mtx sync.Mutex

	done         atomic.Bool
	manager      instance.Manager
	promInstance string

	constLabels labels.Labels
	namespace   string

	logger log.Logger
}

func newRemoteWriteExporter(cfg *Config) (component.MetricsExporter, error) {
	logger := log.With(util.Logger, "component", "traces remote write exporter")

	ls := make(labels.Labels, 0, len(cfg.ConstLabels))

	for name, value := range cfg.ConstLabels {
		ls = append(ls, labels.Label{Name: name, Value: value})
	}

	return &remoteWriteExporter{
		mtx:          sync.Mutex{},
		done:         atomic.Bool{},
		constLabels:  ls,
		namespace:    cfg.Namespace,
		promInstance: cfg.PromInstance,
		logger:       logger,
	}, nil
}

func (e *remoteWriteExporter) Start(ctx context.Context, _ component.Host) error {
	manager, ok := ctx.Value(contextkeys.Metrics).(instance.Manager)
	if !ok || manager == nil {
		return fmt.Errorf("key does not contain a InstanceManager instance")
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

	// Lock taken to ensure that only one appender is open at a time.
	// This prevents parallel writes for metrics with the same labels.
	e.mtx.Lock()
	defer e.mtx.Unlock()

	prom, err := e.manager.GetInstance(e.promInstance)
	if err != nil {
		level.Warn(e.logger).Log("msg", "failed to get prom instance", "err", err)
		return nil
	}
	app := prom.Appender(ctx)

	resourceMetrics := md.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		resourceMetric := resourceMetrics.At(i)
		instrumentationLibraryMetricsSlice := resourceMetric.InstrumentationLibraryMetrics()
		for j := 0; j < instrumentationLibraryMetricsSlice.Len(); j++ {
			metricSlice := instrumentationLibraryMetricsSlice.At(j).Metrics()
			for k := 0; k < metricSlice.Len(); k++ {
				switch metric := metricSlice.At(k); metric.DataType() {
				case pdata.MetricDataTypeGauge:
					dataPoints := metric.Sum().DataPoints()
					if err := e.handleNumberDataPoints(app, metric.Name(), dataPoints); err != nil {
						return err
					}
				case pdata.MetricDataTypeSum:
					if metric.Sum().AggregationTemporality() != pdata.MetricAggregationTemporalityCumulative {
						continue // Only cumulative metrics are supported
					}
					dataPoints := metric.Sum().DataPoints()
					if err := e.handleNumberDataPoints(app, metric.Name(), dataPoints); err != nil {
						return err
					}
				case pdata.MetricDataTypeHistogram:
					if metric.Histogram().AggregationTemporality() != pdata.MetricAggregationTemporalityCumulative {
						continue // Only cumulative metrics are supported
					}
					dataPoints := metric.Histogram().DataPoints()
					if err := e.handleHistogramDataPoints(app, metric.Name(), dataPoints); err != nil {
						return fmt.Errorf("failed to process metric %s", err)
					}
				case pdata.MetricDataTypeSummary:
					return fmt.Errorf("unsupported metric data type %s", metric.DataType())
				default:
					return fmt.Errorf("unsupported metric data type %s", metric.DataType())
				}
			}
		}
	}

	return app.Commit()
}

func (e *remoteWriteExporter) handleNumberDataPoints(app storage.Appender, name string, dataPoints pdata.NumberDataPointSlice) error {
	for ix := 0; ix < dataPoints.Len(); ix++ {
		dataPoint := dataPoints.At(ix)
		lbls := e.createLabelSet(name, noSuffix, dataPoint.Attributes(), labels.Labels{})
		if err := e.appendNumberDataPoint(app, dataPoint, lbls); err != nil {
			return fmt.Errorf("failed to process metric %s", err)
		}
	}
	return nil
}

func (e *remoteWriteExporter) appendNumberDataPoint(app storage.Appender, dataPoint pdata.NumberDataPoint, labels labels.Labels) error {
	var val float64
	switch dataPoint.ValueType() {
	case pdata.MetricValueTypeDouble:
		val = dataPoint.DoubleVal()
	case pdata.MetricValueTypeInt:
		val = float64(dataPoint.IntVal())
	default:
		return fmt.Errorf("unknown data point type: %s", dataPoint.ValueType())
	}
	ts := e.timestamp()

	_, err := app.Append(0, labels, ts, val)
	return err
}

func (e *remoteWriteExporter) handleHistogramDataPoints(app storage.Appender, name string, dataPoints pdata.HistogramDataPointSlice) error {
	for ix := 0; ix < dataPoints.Len(); ix++ {
		dataPoint := dataPoints.At(ix)
		ts := e.timestamp()

		// Append sum value
		sumLabels := e.createLabelSet(name, sumSuffix, dataPoint.Attributes(), labels.Labels{})
		if _, err := app.Append(0, sumLabels, ts, dataPoint.Sum()); err != nil {
			return err
		}

		// Append count value
		countLabels := e.createLabelSet(name, countSuffix, dataPoint.Attributes(), labels.Labels{})
		if _, err := app.Append(0, countLabels, ts, float64(dataPoint.Count())); err != nil {
			return err
		}

		var cumulativeCount uint64
		for ix, eb := range dataPoint.ExplicitBounds() {
			if ix >= len(dataPoint.BucketCounts()) {
				break
			}
			cumulativeCount += dataPoint.BucketCounts()[ix]
			boundStr := strconv.FormatFloat(eb, 'f', -1, 64)
			bucketLabels := e.createLabelSet(name, bucketSuffix, dataPoint.Attributes(), labels.Labels{{Name: leStr, Value: boundStr}})
			if _, err := app.Append(0, bucketLabels, ts, float64(cumulativeCount)); err != nil {
				return err
			}
		}
		// add le=+Inf bucket
		cumulativeCount += dataPoint.BucketCounts()[len(dataPoint.BucketCounts())-1]
		infBucketLabels := e.createLabelSet(name, bucketSuffix, dataPoint.Attributes(), labels.Labels{{Name: leStr, Value: infBucket}})
		if _, err := app.Append(0, infBucketLabels, ts, float64(cumulativeCount)); err != nil {
			return err
		}
	}
	return nil
}

func (e *remoteWriteExporter) createLabelSet(name, suffix string, labelMap pdata.AttributeMap, customLabels labels.Labels) labels.Labels {
	ls := make(labels.Labels, 0, labelMap.Len()+1+len(e.constLabels)+len(customLabels))
	// Labels from spanmetrics processor
	labelMap.Range(func(k string, v pdata.AttributeValue) bool {
		ls = append(ls, labels.Label{
			Name:  strings.Replace(k, ".", "_", -1),
			Value: v.StringVal(),
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

func (e *remoteWriteExporter) timestamp() int64 {
	return time.Now().UnixMilli()
}

func metricName(namespace, metric, suffix string) string {
	if len(suffix) != 0 {
		return fmt.Sprintf("%s_%s_%s", namespace, metric, suffix)
	}
	return fmt.Sprintf("%s_%s", namespace, metric)
}
