package promsdprocessor

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	util "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.6.1"
)

type promServiceDiscoProcessor struct {
	nextConsumer     consumer.Traces
	discoveryMgr     *discovery.Manager
	discoveryMgrStop context.CancelFunc
	discoveryMgrCtx  context.Context

	relabelConfigs map[string][]*relabel.Config
	hostLabels     map[string]model.LabelSet
	mtx            sync.Mutex

	operationType   string
	podAssociations []string

	logger log.Logger
}

func newTraceProcessor(nextConsumer consumer.Traces, operationType string, podAssociations []string, scrapeConfigs []*config.ScrapeConfig) (component.TracesProcessor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	logger := log.With(util.Logger, "component", "traces service disco")
	mgr := discovery.NewManager(ctx, logger, discovery.Name("traces service disco"))

	relabelConfigs := map[string][]*relabel.Config{}
	managerConfig := map[string]discovery.Configs{}
	for _, v := range scrapeConfigs {
		managerConfig[v.JobName] = v.ServiceDiscoveryConfigs
		relabelConfigs[v.JobName] = v.RelabelConfigs
	}

	err := mgr.ApplyConfig(managerConfig)
	if err != nil {
		cancel()
		return nil, err
	}

	switch operationType {
	case OperationTypeUpsert, OperationTypeInsert, OperationTypeUpdate:
	case "": // Use Upsert by default
		operationType = OperationTypeUpsert
	default:
		cancel()
		return nil, fmt.Errorf("unknown operation type %s", operationType)
	}

	for _, podAssociation := range podAssociations {
		switch podAssociation {
		case podAssociationIPLabel, podAssociationOTelIPLabel, podAssociationk8sIPLabel, podAssociationHostnameLabel, podAssociationConnectionIP:
		default:
			cancel()
			return nil, fmt.Errorf("unknown pod association %s", podAssociation)
		}
	}

	if len(podAssociations) == 0 {
		podAssociations = []string{podAssociationIPLabel, podAssociationOTelIPLabel, podAssociationk8sIPLabel, podAssociationHostnameLabel, podAssociationConnectionIP}
	}

	if nextConsumer == nil {
		cancel()
		return nil, component.ErrNilNextConsumer
	}
	return &promServiceDiscoProcessor{
		nextConsumer:     nextConsumer,
		discoveryMgr:     mgr,
		discoveryMgrStop: cancel,
		discoveryMgrCtx:  ctx,
		relabelConfigs:   relabelConfigs,
		hostLabels:       make(map[string]model.LabelSet),
		logger:           logger,
		operationType:    operationType,
		podAssociations:  podAssociations,
	}, nil
}

func (p *promServiceDiscoProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)

		p.processAttributes(ctx, rs.Resource().Attributes())
	}

	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func stringAttributeFromMap(attrs pcommon.Map, key string) string {
	if attr, ok := attrs.Get(key); ok {
		if attr.Type() == pcommon.ValueTypeStr {
			return attr.Str()
		}
	}
	return ""
}

func (p *promServiceDiscoProcessor) getConnectionIP(ctx context.Context) string {
	c := client.FromContext(ctx)
	if c.Addr == nil {
		return ""
	}

	host := c.Addr.String()
	if strings.Contains(host, ":") {
		var err error
		splitHost, _, err := net.SplitHostPort(host)
		if err != nil {
			// It's normal for this to fail for IPv6 address strings that don't actually include a port.
			level.Debug(p.logger).Log("msg", "unable to split connection host and port", "host", host, "err", err)
		} else {
			host = splitHost
		}
	}

	return host
}

func (p *promServiceDiscoProcessor) getPodIP(ctx context.Context, attrs pcommon.Map) string {
	for _, podAssociation := range p.podAssociations {
		switch podAssociation {
		case podAssociationIPLabel, podAssociationOTelIPLabel, podAssociationk8sIPLabel:
			ip := stringAttributeFromMap(attrs, podAssociation)
			if ip != "" {
				return ip
			}
		case podAssociationHostnameLabel:
			hostname := stringAttributeFromMap(attrs, semconv.AttributeHostName)
			if net.ParseIP(hostname) != nil {
				return hostname
			}
		case podAssociationConnectionIP:
			ip := p.getConnectionIP(ctx)
			if ip != "" {
				return ip
			}
		}
	}
	return ""
}

func (p *promServiceDiscoProcessor) processAttributes(ctx context.Context, attrs pcommon.Map) {
	ip := p.getPodIP(ctx, attrs)
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
		switch p.operationType {
		case OperationTypeUpsert:
			attrs.PutStr(string(k), string(v))
		case OperationTypeInsert:
			if _, ok := attrs.Get(string(k)); !ok {
				attrs.PutStr(string(k), string(v))
			}
		case OperationTypeUpdate:
			if toVal, ok := attrs.Get(string(k)); ok {
				toVal.SetStr(string(v))
			}
		}
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
		if processedLabels == nil { // dropped
			continue
		}

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
