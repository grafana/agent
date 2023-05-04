package vsphere

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/sync/semaphore"
)

type vsphereCollector struct {
	logger   log.Logger
	endpoint *endpoint
	sem      *semaphore.Weighted
}

func (c *vsphereCollector) Describe(chan<- *prometheus.Desc) {
	level.Debug(c.logger).Log("msg", "describe")
}

func (c *vsphereCollector) Collect(metrics chan<- prometheus.Metric) {
	ctx := context.Background()
	myClient, err := c.endpoint.clientFactory.GetClient(ctx)
	if err != nil {
		level.Debug(c.logger).Log("msg", "error getting client", "err", err)
		return
	}

	if c.endpoint.cfg.ObjectDiscoveryInterval == 0 {
		err := c.endpoint.discover(ctx)
		if err != nil && err != context.Canceled {
			level.Error(c.logger).Log("msg", "discovery error", "host", c.endpoint.url.Host, "err", err.Error())
			return
		}
	}

	c.endpoint.collectMux.RLock()
	defer c.endpoint.collectMux.RUnlock()

	now, err := myClient.getServerTime(ctx)
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to get server time", "err", err.Error())
		return
	}

	var wg sync.WaitGroup
	for k, r := range c.endpoint.resourceKinds {
		if r.enabled {
			level.Debug(c.logger).Log("msg", "collecting metrics", "kind", k)
			wg.Add(1)
			go func(kind string, res *resourceKind) {
				defer wg.Done()
				c.collectResource(ctx, metrics, now, myClient, kind, res)
			}(k, r)
		}
	}
	wg.Wait()
}

func (c *vsphereCollector) collectResource(ctx context.Context, metrics chan<- prometheus.Metric,
	now time.Time, cli *client, kind string, res *resourceKind) {

	latest := res.latestSample
	if !latest.IsZero() {
		elapsed := now.Sub(latest).Seconds() + 5.0 // Allow 5 second jitter.
		if !res.realTime && elapsed < float64(res.sampling) {
			// No new data would be available. We're outta here!
			level.Debug(c.logger).Log("msg", "sampling period has not elapsed", "resource", kind)
			return
		}
	} else {
		latest = now.Add(time.Duration(-res.sampling) * time.Second)
	}

	var refs []types.ManagedObjectReference
	start := latest.Add(time.Duration(-res.sampling) * time.Second * (time.Duration(c.endpoint.cfg.MetricLookback) - 1))
	for _, obj := range res.objects {
		refs = append(refs, obj.ref)
	}
	level.Debug(c.logger).Log("refs count", len(refs), "kind", kind)

	spec := types.PerfQuerySpec{
		MaxSample:  1,
		MetricId:   []types.PerfMetricId{{Instance: ""}},
		IntervalId: res.sampling,
		StartTime:  &start,
		EndTime:    &now,
	}

	// chunk refs and collect
	var (
		ccWg         sync.WaitGroup
		refsSize     = len(refs)
		latestSample = time.Time{}
		chunkSize    = c.endpoint.cfg.RefChunkSize
		latestMut    sync.RWMutex
	)
	for i := 0; i < refsSize; i += chunkSize {
		end := i + chunkSize
		if end > refsSize {
			end = refsSize
		}
		ccWg.Add(1)
		go func(chunk []types.ManagedObjectReference, pRes *resourceKind) {
			defer ccWg.Done()
			latestMut.RLock()
			if sampleTime := c.collectChunk(ctx, metrics, cli, spec, chunk, pRes); sampleTime != nil &&
				sampleTime.After(latestSample) && !sampleTime.IsZero() {

				latestMut.RUnlock()
				latestMut.Lock()
				latestSample = *sampleTime
				latestMut.Unlock()
			} else {
				latestMut.RUnlock()
			}
		}(refs[i:end], res)
	}
	ccWg.Wait()
	latestMut.RLock()
	if !latestSample.IsZero() {
		res.latestSample = latestSample
	}
	latestMut.RUnlock()
}

func (c *vsphereCollector) collectChunk(ctx context.Context, metrics chan<- prometheus.Metric, cli *client,
	spec types.PerfQuerySpec, chunk []types.ManagedObjectReference, res *resourceKind) *time.Time {

	defer func() {
		c.sem.Release(1)
	}()
	if err := c.sem.Acquire(ctx, 1); err != nil {
		level.Error(c.logger).Log("msg", "error acquiring semaphore", "err", err)
		return nil
	}
	sampleTime, err := c.collect(ctx, cli, spec, metrics, chunk, res)
	if err != nil {
		level.Error(c.logger).Log("msg", "error collecting chunk", "err", err)
		return nil
	}
	return sampleTime
}

func (c *vsphereCollector) collect(ctx context.Context, cli *client, spec types.PerfQuerySpec,
	metrics chan<- prometheus.Metric, chunk []types.ManagedObjectReference, res *resourceKind) (*time.Time, error) {

	counters, err := cli.counterInfoByName(ctx)
	if err != nil {
		level.Error(c.logger).Log("msg", "error getting counters", "err", err)
		return nil, err
	}

	var names []string
	for name := range counters {
		names = append(names, name)
	}
	sample, err := cli.Perf.SampleByName(ctx, spec, names, chunk)
	if err != nil {
		level.Error(c.logger).Log("msg", "error getting sample by name", "err", err)
		return nil, err
	}

	result, err := cli.Perf.ToMetricSeries(ctx, sample)
	if err != nil {
		level.Error(c.logger).Log("err", err)
		return nil, err
	}

	var (
		parent     string
		parentType string
	)

	for _, metric := range result {
		mo := strings.Split(metric.Entity.String(), ":")[1]

		constLabels := make(prometheus.Labels)
		constLabels["moid"] = mo
		constLabels["name"] = c.endpoint.resourceKinds[res.name].objects[mo].name

		// add type/parent labels
		parent = c.endpoint.resourceKinds[res.name].objects[mo].parentRef.Value
		parentType = res.parent
		for parent != "" {
			if pRes, ok := c.endpoint.resourceKinds[parentType]; ok {
				if pObj := pRes.objects[parent]; pObj != nil {
					constLabels[parentType] = pObj.name
					parent = c.endpoint.resourceKinds[pRes.name].objects[parent].parentRef.Value
					parentType = pRes.parent
					continue
				}
			}
			parent = ""
			parentType = ""
		}

		for _, v := range metric.Value {
			counter := counters[v.Name]
			units := counter.UnitInfo.GetElementDescription().Label
			if len(v.Value) != 0 {
				// get fqName
				fqName := fmt.Sprintf("vsphere_%s_%s", metric.Entity.Type, strings.ReplaceAll(v.Name, ".", "_"))

				desc := prometheus.NewDesc(
					fqName, fmt.Sprintf("metric: %s units: %s", v.Name, units),
					nil,
					constLabels)

				// send metric, using v.Value[0] since we're only requesting a single sample at this time.
				// TODO: need to make sure that this is what we want to do here -- in some cases vsphere is returning
				// multiple samples for a counter because there are multiple instances of the resource e.g. cpu cores
				m, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, float64(v.Value[0]))
				if err != nil {
					level.Error(c.logger).Log("err", err)
					continue
				}
				metrics <- m
			}
		}
	}
	now := time.Now()
	return &now, nil
}

func newVSphereCollector(ctx context.Context, logger log.Logger, e *endpoint) (prometheus.Collector, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	err := e.init(ctx)
	if err != nil {
		return nil, err
	}

	return &vsphereCollector{
		logger:   logger,
		endpoint: e,
		sem:      semaphore.NewWeighted(int64(e.cfg.CollectConcurrency)),
	}, nil
}
