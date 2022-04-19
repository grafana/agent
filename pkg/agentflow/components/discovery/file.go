package discovery

import (
	"context"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	"github.com/iancoleman/orderedmap"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/file"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"time"
)

type File struct {
	self       *actor.PID
	outs       []*actor.PID
	name       string
	disc       *file.Discovery
	targetCh   chan []*targetgroup.Group
	cancelDisc context.CancelFunc
	log        log.Logger
	root       *actor.RootContext
}

func NewFile(name string, cfg *config.FileServiceDiscovery, global types.Global) (actorstate.FlowActor, error) {
	if len(cfg.Locations) == 0 {
		return nil, fmt.Errorf("unable to create file discovery due to missing locations")
	}
	if cfg.RefreshInterval == 0 {
		cfg.RefreshInterval = model.Duration(5 * time.Minute)
	}
	//TODO add check to ensure each location is legit
	c := &file.SDConfig{
		Files:           cfg.Locations,
		RefreshInterval: cfg.RefreshInterval,
	}
	disc := file.NewDiscovery(c, global.Log)

	return &File{
		name:     name,
		disc:     disc,
		log:      global.Log,
		targetCh: make(chan []*targetgroup.Group),
	}, nil
}

func (f *File) PID() *actor.PID {
	return f.self
}

func (f *File) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.None}
}

func (f *File) Output() actorstate.InOutType {
	return actorstate.Targets
}

func (f *File) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Init:
		f.outs = msg.Children
	case actorstate.Start:
		f.self = c.Self()
		ctx := context.Background()
		cancellable, cancelFunc := context.WithCancel(ctx)
		f.cancelDisc = cancelFunc
		go f.disc.Run(cancellable, f.targetCh)
	case actorstate.Stop:
		f.cancelDisc()
	case []*targetgroup.Group:
		f.handleTargets(msg, c)
	}
}

func (f *File) Name() string {
	return f.name
}

func (f *File) handleChannel() {
	for {
		select {
		case targets := <-f.targetCh:
			// Call self so it gets queued and we dont have to worry about concurrency
			f.root.Send(f.self, targets)
		}
	}
}

func (f *File) handleTargets(in []*targetgroup.Group, c actor.Context) {
	targets := make([]exchange.Target, 0)
	for _, tg := range in {
		grpTargets := f.newTargetsFromGroup(tg)
		targets = append(targets, grpTargets...)
	}
	ts := exchange.NewTargetSet(f.name, targets)
	for _, o := range f.outs {
		c.Send(o, ts)
	}
}

func (f *File) newTargetsFromGroup(group *targetgroup.Group) []exchange.Target {
	returnTargets := make([]exchange.Target, 0)
	for _, t := range group.Targets {
		state := exchange.New
		nt := exchange.NewTarget(string(t["__address__"]), group.Source, exchange.CopyLabelSet(group.Labels), orderedmap.New(), state)
		returnTargets = append(returnTargets, nt)
	}
	return returnTargets
}
