package promsdprocessor

import (
	"context"
	"net"
	"strings"
	"sync"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/relabel"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
)

type promServiceDiscoProcessor struct {
	nextConsumer     consumer.Traces
	discoveryMgr     *discovery.Manager
	discoveryMgrStop context.CancelFunc
	discoveryMgrCtx  context.Context

	relabelConfigs map[string][]*relabel.Config
	hostLabels     map[string]model.LabelSet
	mtx            sync.Mutex

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.Traces, scrapeConfigs []*config.ScrapeConfig) (component.TracesProcessor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.With(util.Logger, "component", "tempo service disco")
	mgr := discovery.NewManager(ctx, logger, discovery.Name("tempo service disco"))

	relabelConfigs := map[string][]*relabel.Config{}
	cfg := map[string]discovery.Configs{}
	for _, v := range scrapeConfigs {
		cfg[v.JobName] = v.ServiceDiscoveryConfigs
		relabelConfigs[v.JobName] = v.RelabelConfigs
	}

	err := mgr.ApplyConfig(cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	if nextConsumer == nil {
		cancel()
		return nil, componenterror.ErrNilNextConsumer
	}
	return &promServiceDiscoProcessor{
		nextConsumer:     nextConsumer,
		discoveryMgr:     mgr,
		discoveryMgrStop: cancel,
		discoveryMgrCtx:  ctx,
		relabelConfigs:   relabelConfigs,
		hostLabels:       make(map[string]model.LabelSet),
		logger:           logger,
	}, nil
}

func (p *promServiceDiscoProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)

		p.processAttributes(rs.Resource().Attributes())
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (p *promServiceDiscoProcessor) processAttributes(attrs pdata.AttributeMap) {
	// find the ip
	ipTagNames := []string{
		"ip",          // jaeger/opentracing? default
		"net.host.ip", // otel semantics for host ip
	}

	var ip string
	for _, name := range ipTagNames {
		val, ok := attrs.Get(name)
		if !ok {
			continue
		}

		ip = val.StringVal()
		break
	}

	// have to have an ip for labels lookup
	if ip == "" {
		level.Debug(p.logger).Log("msg", "unable to find ip in span attributes, skipping attribute addition")
		return
	}

	p.mtx.Lock()
	defer p.mtx.Unlock()

	labels, ok := p.hostLabels[ip]
	if !ok {
		level.Debug(p.logger).Log("msg", "unable to find matching hostLabels", "ip", ip)
		return
	}

	for k, v := range labels {
		attrs.UpsertString(string(k), string(v))
	}
}

func (p *promServiceDiscoProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// Start is invoked during service startup.
func (p *promServiceDiscoProcessor) Start(_ context.Context, _ component.Host) error {
	go p.watchServiceDiscovery()

	go func() {
		err := p.discoveryMgr.Run()
		if err != nil && err != context.Canceled {
			level.Error(p.logger).Log("msg", "failed to start prom svc disco.  relabeling disabled", "err", err)
		}
	}()

	return nil
}

// Shutdown is invoked during service shutdown.
func (p *promServiceDiscoProcessor) Shutdown(context.Context) error {
	if p.discoveryMgrStop != nil {
		p.discoveryMgrStop()
	}
	return nil
}

func (p *promServiceDiscoProcessor) watchServiceDiscovery() {
	for {
		// p.discoveryMgr.SyncCh() is never closed so we need to watch the context as well to properly exit this goroutine
		select {
		case targetGroups := <-p.discoveryMgr.SyncCh():
			hostLabels := make(map[string]model.LabelSet)
			level.Debug(p.logger).Log("msg", "syncing target groups", "count", len(targetGroups))
			for jobName, groups := range targetGroups {
				p.syncGroups(jobName, groups, hostLabels)
			}
			p.mtx.Lock()
			p.hostLabels = hostLabels
			p.mtx.Unlock()
		case <-p.discoveryMgrCtx.Done():
			return
		}
	}
}

func (p *promServiceDiscoProcessor) syncGroups(jobName string, groups []*targetgroup.Group, hostLabels map[string]model.LabelSet) {
	level.Debug(p.logger).Log("msg", "syncing target group", "jobName", jobName)
	for _, g := range groups {
		p.syncTargets(jobName, g, hostLabels)
	}
}

func (p *promServiceDiscoProcessor) syncTargets(jobName string, group *targetgroup.Group, hostLabels map[string]model.LabelSet) {
	level.Debug(p.logger).Log("msg", "syncing targets", "count", len(group.Targets))

	relabelConfig := p.relabelConfigs[jobName]
	if relabelConfig == nil {
		level.Warn(p.logger).Log("msg", "relabel config not found for job. skipping labeling", "jobName", jobName)
		return
	}

	for _, t := range group.Targets {
		discoveredLabels := group.Labels.Merge(t)

		level.Debug(p.logger).Log("discoveredLabels", discoveredLabels)
		var labelMap = make(map[string]string)
		for k, v := range discoveredLabels.Clone() {
			labelMap[string(k)] = string(v)
		}
		processedLabels := relabel.Process(labels.FromMap(labelMap), relabelConfig...)
		level.Debug(p.logger).Log("processedLabels", processedLabels)

		var labels = make(model.LabelSet)
		for k, v := range processedLabels.Map() {
			labels[model.LabelName(k)] = model.LabelValue(v)
		}

		address, ok := labels[model.AddressLabel]
		if !ok {
			level.Warn(p.logger).Log("msg", "ignoring target, unable to find address", "labels", labels.String())
			continue
		}

		host := string(address)
		if strings.Contains(host, ":") {
			var err error
			host, _, err = net.SplitHostPort(host)
			if err != nil {
				level.Warn(p.logger).Log("msg", "unable to split host port", "address", address, "err", err)
				continue
			}
		}

		for k := range labels {
			if strings.HasPrefix(string(k), "__") {
				delete(labels, k)
			}
		}

		level.Debug(p.logger).Log("msg", "adding host to hostLabels", "host", host)
		hostLabels[host] = labels
	}
}
