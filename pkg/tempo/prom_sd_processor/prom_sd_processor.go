package prom_sd_processor

import (
	"context"
	"fmt"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	sd_config "github.com/prometheus/prometheus/discovery/config"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
)

type promServiceDiscoProcessor struct {
	nextConsumer consumer.TraceConsumer
	discoveryMgr *discovery.Manager
}

func newTraceProcessor(nextConsumer consumer.TraceConsumer, scrapeConfigs []*config.ScrapeConfig) (component.TraceProcessor, error) {
	logger := log.With(util.Logger, "component", "tempo service disco")                              // jpe i.logger?
	mgr := discovery.NewManager(context.Background(), logger, discovery.Name("tempo service disco")) // jpe ?

	cfg := map[string]sd_config.ServiceDiscoveryConfig{}
	for _, v := range scrapeConfigs {
		cfg[v.JobName] = v.ServiceDiscoveryConfig
	}

	err := mgr.ApplyConfig(cfg)
	if err != nil {
		return nil, err
	}

	if nextConsumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}
	return &promServiceDiscoProcessor{
		nextConsumer: nextConsumer,
		discoveryMgr: mgr,
	}, nil
}

func (p *promServiceDiscoProcessor) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	return p.nextConsumer.ConsumeTraces(ctx, td)
}

func (a *promServiceDiscoProcessor) GetCapabilities() component.ProcessorCapabilities {
	return component.ProcessorCapabilities{MutatesConsumedData: true}
}

// Start is invoked during service startup.
func (p *promServiceDiscoProcessor) Start(_ context.Context, _ component.Host) error {
	err := p.discoveryMgr.Run()
	if err != nil {
		return err
	}

	go p.watchServiceDiscovery()

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

func (p *promServiceDiscoProcessor) syncGroups(jobName string, groups []*targetgroup.Group) { // jpe jobName?
	for _, g := range groups {
		// jpe ? wut
		fmt.Println(g.Source)
	}
}
