// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver/internal"

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.uber.org/zap"
)

const (
	targetMetricName  = "target_info"
	scopeMetricName   = "otel_scope_info"
	scopeNameLabel    = "otel_scope_name"
	scopeVersionLabel = "otel_scope_version"
	receiverName      = "otelcol/prometheusreceiver"
)

type transaction struct {
	isNew           bool
	trimSuffixes    bool
	ctx             context.Context
	families        map[scopeID]map[string]*metricFamily
	mc              scrape.MetricMetadataStore
	sink            consumer.Metrics
	externalLabels  labels.Labels
	nodeResource    pcommon.Resource
	scopeAttributes map[scopeID]pcommon.Map
	logger          *zap.Logger
	buildInfo       component.BuildInfo
	metricAdjuster  MetricsAdjuster
	obsrecv         *receiverhelper.ObsReport
	// Used as buffer to calculate series ref hash.
	bufBytes []byte
}

var emptyScopeID scopeID

type scopeID struct {
	name    string
	version string
}

func newTransaction(
	ctx context.Context,
	metricAdjuster MetricsAdjuster,
	sink consumer.Metrics,
	externalLabels labels.Labels,
	settings receiver.CreateSettings,
	obsrecv *receiverhelper.ObsReport,
	trimSuffixes bool) *transaction {

	return &transaction{
		ctx:             ctx,
		families:        make(map[scopeID]map[string]*metricFamily),
		isNew:           true,
		trimSuffixes:    trimSuffixes,
		sink:            sink,
		metricAdjuster:  metricAdjuster,
		externalLabels:  externalLabels,
		logger:          settings.Logger,
		buildInfo:       settings.BuildInfo,
		obsrecv:         obsrecv,
		bufBytes:        make([]byte, 0, 1024),
		scopeAttributes: make(map[scopeID]pcommon.Map),
	}
}

// Append always returns 0 to disable label caching.
func (t *transaction) Append(_ storage.SeriesRef, ls labels.Labels, atMs int64, val float64) (storage.SeriesRef, error) {
	select {
	case <-t.ctx.Done():
		return 0, errTransactionAborted
	default:
	}

	if len(t.externalLabels) != 0 {
		ls = append(ls, t.externalLabels...)
		sort.Sort(ls)
	}

	if t.isNew {
		if err := t.initTransaction(ls); err != nil {
			return 0, err
		}
	}

	// Any datapoint with duplicate labels MUST be rejected per:
	// * https://github.com/open-telemetry/wg-prometheus/issues/44
	// * https://github.com/open-telemetry/opentelemetry-collector/issues/3407
	// as Prometheus rejects such too as of version 2.16.0, released on 2020-02-13.
	if dupLabel, hasDup := ls.HasDuplicateLabelNames(); hasDup {
		return 0, fmt.Errorf("invalid sample: non-unique label names: %q", dupLabel)
	}

	metricName := ls.Get(model.MetricNameLabel)
	if metricName == "" {
		return 0, errMetricNameNotFound
	}

	// See https://www.prometheus.io/docs/concepts/jobs_instances/#automatically-generated-labels-and-time-series
	// up: 1 if the instance is healthy, i.e. reachable, or 0 if the scrape failed.
	// But it can also be a staleNaN, which is inserted when the target goes away.
	if metricName == scrapeUpMetricName && val != 1.0 && !value.IsStaleNaN(val) {
		if val == 0.0 {
			t.logger.Warn("Failed to scrape Prometheus endpoint",
				zap.Int64("scrape_timestamp", atMs),
				zap.Stringer("target_labels", ls))
		} else {
			t.logger.Warn("The 'up' metric contains invalid value",
				zap.Float64("value", val),
				zap.Int64("scrape_timestamp", atMs),
				zap.Stringer("target_labels", ls))
		}
	}

	// For the `target_info` metric we need to convert it to resource attributes.
	if metricName == targetMetricName {
		t.AddTargetInfo(ls)
		return 0, nil
	}

	// For the `otel_scope_info` metric we need to convert it to scope attributes.
	if metricName == scopeMetricName {
		t.addScopeInfo(ls)
		return 0, nil
	}

	curMF := t.getOrCreateMetricFamily(getScopeID(ls), metricName)
	err := curMF.addSeries(t.getSeriesRef(ls, curMF.mtype), metricName, ls, atMs, val)
	if err != nil {
		t.logger.Warn("failed to add datapoint", zap.Error(err), zap.String("metric_name", metricName), zap.Any("labels", ls))
	}

	return 0, nil // never return errors, as that fails the whole scrape
}

func (t *transaction) getOrCreateMetricFamily(scope scopeID, mn string) *metricFamily {
	_, ok := t.families[scope]
	if !ok {
		t.families[scope] = make(map[string]*metricFamily)
	}
	curMf, ok := t.families[scope][mn]
	if !ok {
		fn := mn
		if _, ok := t.mc.GetMetadata(mn); !ok {
			fn = normalizeMetricName(mn)
		}
		if mf, ok := t.families[scope][fn]; ok && mf.includesMetric(mn) {
			curMf = mf
		} else {
			curMf = newMetricFamily(mn, t.mc, t.logger)
			t.families[scope][curMf.name] = curMf
		}
	}
	return curMf
}

func (t *transaction) AppendExemplar(_ storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	select {
	case <-t.ctx.Done():
		return 0, errTransactionAborted
	default:
	}

	if t.isNew {
		if err := t.initTransaction(l); err != nil {
			return 0, err
		}
	}

	l = l.WithoutEmpty()

	if dupLabel, hasDup := l.HasDuplicateLabelNames(); hasDup {
		return 0, fmt.Errorf("invalid sample: non-unique label names: %q", dupLabel)
	}

	mn := l.Get(model.MetricNameLabel)
	if mn == "" {
		return 0, errMetricNameNotFound
	}

	mf := t.getOrCreateMetricFamily(getScopeID(l), mn)
	mf.addExemplar(t.getSeriesRef(l, mf.mtype), e)

	return 0, nil
}

func (t *transaction) AppendHistogram(_ storage.SeriesRef, _ labels.Labels, _ int64, _ *histogram.Histogram, _ *histogram.FloatHistogram) (storage.SeriesRef, error) {
	//TODO: implement this func
	return 0, nil
}

func (t *transaction) getSeriesRef(ls labels.Labels, mtype pmetric.MetricType) uint64 {
	var hash uint64
	hash, t.bufBytes = getSeriesRef(t.bufBytes, ls, mtype)
	return hash
}

// getMetrics returns all metrics to the given slice.
// The only error returned by this function is errNoDataToBuild.
func (t *transaction) getMetrics(resource pcommon.Resource) (pmetric.Metrics, error) {
	if len(t.families) == 0 {
		return pmetric.Metrics{}, errNoDataToBuild
	}

	md := pmetric.NewMetrics()
	rms := md.ResourceMetrics().AppendEmpty()
	resource.CopyTo(rms.Resource())

	for scope, mfs := range t.families {
		ils := rms.ScopeMetrics().AppendEmpty()
		// If metrics don't include otel_scope_name or otel_scope_version
		// labels, use the receiver name and version.
		if scope == emptyScopeID {
			ils.Scope().SetName(receiverName)
			ils.Scope().SetVersion(t.buildInfo.Version)
		} else {
			// Otherwise, use the scope that was provided with the metrics.
			ils.Scope().SetName(scope.name)
			ils.Scope().SetVersion(scope.version)
			// If we got an otel_scope_info metric for that scope, get scope
			// attributes from it.
			attributes, ok := t.scopeAttributes[scope]
			if ok {
				attributes.CopyTo(ils.Scope().Attributes())
			}
		}
		metrics := ils.Metrics()
		for _, mf := range mfs {
			mf.appendMetric(metrics, t.trimSuffixes)
		}
	}

	return md, nil
}

func getScopeID(ls labels.Labels) scopeID {
	var scope scopeID
	for _, lbl := range ls {
		if lbl.Name == scopeNameLabel {
			scope.name = lbl.Value
		}
		if lbl.Name == scopeVersionLabel {
			scope.version = lbl.Value
		}
	}
	return scope
}

func (t *transaction) initTransaction(labels labels.Labels) error {
	target, ok := scrape.TargetFromContext(t.ctx)
	if !ok {
		return errors.New("unable to find target in context")
	}
	t.mc, ok = scrape.MetricMetadataStoreFromContext(t.ctx)
	if !ok {
		return errors.New("unable to find MetricMetadataStore in context")
	}

	job, instance := labels.Get(model.JobLabel), labels.Get(model.InstanceLabel)
	if job == "" || instance == "" {
		return errNoJobInstance
	}
	t.nodeResource = CreateResource(job, instance, target.DiscoveredLabels())
	t.isNew = false
	return nil
}

func (t *transaction) Commit() error {
	if t.isNew {
		return nil
	}

	ctx := t.obsrecv.StartMetricsOp(t.ctx)
	md, err := t.getMetrics(t.nodeResource)
	if err != nil {
		t.obsrecv.EndMetricsOp(ctx, dataformat, 0, err)
		return err
	}

	numPoints := md.DataPointCount()
	if numPoints == 0 {
		return nil
	}

	if err = t.metricAdjuster.AdjustMetrics(md); err != nil {
		t.obsrecv.EndMetricsOp(ctx, dataformat, numPoints, err)
		return err
	}

	err = t.sink.ConsumeMetrics(ctx, md)
	t.obsrecv.EndMetricsOp(ctx, dataformat, numPoints, err)
	return err
}

func (t *transaction) Rollback() error {
	return nil
}

func (t *transaction) UpdateMetadata(_ storage.SeriesRef, _ labels.Labels, _ metadata.Metadata) (storage.SeriesRef, error) {
	//TODO: implement this func
	return 0, nil
}

func (t *transaction) AddTargetInfo(labels labels.Labels) {
	attrs := t.nodeResource.Attributes()
	for _, lbl := range labels {
		if lbl.Name == model.JobLabel || lbl.Name == model.InstanceLabel || lbl.Name == model.MetricNameLabel {
			continue
		}
		attrs.PutStr(lbl.Name, lbl.Value)
	}
}

func (t *transaction) addScopeInfo(labels labels.Labels) {
	attrs := pcommon.NewMap()
	scope := scopeID{}
	for _, lbl := range labels {
		if lbl.Name == model.JobLabel || lbl.Name == model.InstanceLabel || lbl.Name == model.MetricNameLabel {
			continue
		}
		if lbl.Name == scopeNameLabel {
			scope.name = lbl.Value
			continue
		}
		if lbl.Name == scopeVersionLabel {
			scope.version = lbl.Value
			continue
		}
		attrs.PutStr(lbl.Name, lbl.Value)
	}
	t.scopeAttributes[scope] = attrs
}

func getSeriesRef(bytes []byte, ls labels.Labels, mtype pmetric.MetricType) (uint64, []byte) {
	return ls.HashWithoutLabels(bytes, getSortedNotUsefulLabels(mtype)...)
}
