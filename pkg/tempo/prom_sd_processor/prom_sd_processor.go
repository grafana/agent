package prom_sd_processor

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/relabel"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
)

const (
	addressLabel = "__address__"
)

type promServiceDiscoProcessor struct {
	nextConsumer consumer.TraceConsumer
	discoveryMgr *discovery.Manager

	relabelConfigs map[string][]*relabel.Config
	hostLabels     map[string]model.LabelSet
	mtx            sync.Mutex

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.TraceConsumer, scrapeConfigs []*config.ScrapeConfig) (component.TraceProcessor, error) {
	logger := log.With(util.Logger, "component", "tempo service disco")
	mgr := discovery.NewManager(context.Background(), logger, discovery.Name("tempo service disco"))

	relabelConfigs := map[string][]*relabel.Config{}
	cfg := map[string]sd_config.ServiceDiscoveryConfig{}
	for _, v := range scrapeConfigs {
		cfg[v.JobName] = v.ServiceDiscoveryConfig
		relabelConfigs[v.JobName] = v.RelabelConfigs
	}

	err := mgr.ApplyConfig(cfg)
	if err != nil {
		return nil, err
	}

	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}
	return &promServiceDiscoProcessor{
		nextConsumer:   nextConsumer,
		discoveryMgr:   mgr,
		relabelConfigs: relabelConfigs,
		hostLabels:     make(map[string]model.LabelSet),
		logger:         util.Logger,
	}, nil
}

func (p *promServiceDiscoProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		if rs.IsNil() {
			continue
		}

		ilss := rs.InstrumentationLibrarySpans()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			if ils.IsNil() {
				continue
			}

			ss := ils.Spans()
			for k := 0; k < ss.Len(); k++ {
				s := ss.At(k)
				if s.IsNil() {
					continue
				}

				p.processAttributes(s.Attributes())
			}
		}
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (a *promServiceDiscoProcessor) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{MutatesConsumedData: true}
}

// Start is invoked during service startup.
func (p *promServiceDiscoProcessor) Start(_ context.Context, _ component.Host) error {
	go p.watchServiceDiscovery()

	go func() {
		err := p.discoveryMgr.Run()
		if err != nil {
			level.Error(p.logger).Log("msg", "failed to start prom svc disco.  relabeling disabled", "err", err)
		}
	}()

	return nil
}

// Shutdown is invoked during service shutdown.
func (p *promServiceDiscoProcessor) Shutdown(context.Context) error {
	// jpe - shutdown mgr?
	return nil
}

func (p *promServiceDiscoProcessor) watchServiceDiscovery() {
	for targetGoups := range p.discoveryMgr.SyncCh() {
		for jobName, groups := range targetGoups {
			p.syncGroups(jobName, groups)
		}
	}
}

func (p *promServiceDiscoProcessor) syncGroups(jobName string, groups []*targetgroup.Group) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// wipe and rebuild hostLabels
	p.hostLabels = make(map[string]model.LabelSet)
	for _, g := range groups {
		p.syncTargets(jobName, g)
	}
}

func (p *promServiceDiscoProcessor) syncTargets(jobName string, group *targetgroup.Group) {
	relabelConfig := p.relabelConfigs[jobName]
	if relabelConfig == nil {
		level.Warn(p.logger).Log("msg", "relabel config not found for job. skipping labeling", "jobName", jobName)
		return
	}

	for _, t := range group.Targets {
		discoveredLabels := group.Labels.Merge(t)

		var labelMap = make(map[string]string)
		for k, v := range discoveredLabels.Clone() {
			labelMap[string(k)] = string(v)
		}
		processedLabels := relabel.Process(labels.FromMap(labelMap), relabelConfig...)

		var labels = make(model.LabelSet)
		for k, v := range processedLabels.Map() {
			labels[model.LabelName(k)] = model.LabelValue(v)
		}

		address, ok := labels[addressLabel]
		if !ok {
			level.Warn(p.logger).Log("msg", "ignoring target, unable to find address", "labels", labels.String())
			continue
		}

		host, _, err := net.SplitHostPort(string(address))
		if err != nil {
			level.Warn(p.logger).Log("msg", "unable to split host port", "address", address, "err", err)
			continue
		}

		for k := range labels {
			if strings.HasPrefix(string(k), "__") {
				delete(labels, k)
			}
		}

		p.hostLabels[host] = labels
	}
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
		return
	}

	p.mtx.Lock()
	defer p.mtx.Unlock()

	labels, ok := p.hostLabels[ip]
	if !ok {
		return
	}
	for k, v := range labels {
		attrs.UpsertString(string(k), string(v))
	}
}
