// Package observer implements the Observer service for Flow.
package observer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/agentstate"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/encoding/riverjson"
	"github.com/grafana/agent/service"
)

// ServiceName defines the name used for the Observer service.
const ServiceName = "observer"

type Arguments struct {
	RefreshFrequency time.Duration `river:"refresh_frequency,attr,optional"`
	RemoteEndpoint   string        `river:"remote_endpoint,attr"`
}

var _ river.Defaulter = (*Arguments)(nil)

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	RefreshFrequency: time.Minute,
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

type Observer struct {
	log log.Logger

	mtx          sync.Mutex
	args         Arguments
	configUpdate chan struct{}

	client *agentstate.ParquetClient
}

var _ service.Service = (*Observer)(nil)

// New returns a new, unstarted instance of the HTTP service.
func New(l log.Logger) *Observer {
	return &Observer{
		log:          l,
		configUpdate: make(chan struct{}, 1),

		client: agentstate.NewParquetClient(
			agentstate.NewAgentState(nil),
			nil,
		),
	}
}

// Data implements service.Service.
func (*Observer) Data() any {
	return nil
}

// Definition implements service.Service.
func (*Observer) Definition() service.Definition {
	return service.Definition{
		Name:       ServiceName,
		ConfigType: Arguments{},
		DependsOn:  nil, // observer has no dependencies.
	}
}

// Run implements service.Service.
func (o *Observer) Run(ctx context.Context, host service.Host) error {
	o.observe(ctx, host)

	for {
		o.mtx.Lock()
		refreshFrequency := o.args.RefreshFrequency
		o.mtx.Unlock()

		level.Debug(o.log).Log("msg", "waiting for next refresh before sending state payload", "refresh_frequency", refreshFrequency)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(refreshFrequency):
			o.observe(ctx, host)
		case <-o.configUpdate: // no-op
		}
	}
}

func (o *Observer) observe(ctx context.Context, host service.Host) {
	level.Info(o.log).Log("msg", "sending state payload to remote server")

	o.mtx.Lock()
	defer o.mtx.Unlock()

	rawComponents := component.GetAllComponents(host, component.InfoOptions{
		GetHealth:    true,
		GetArguments: true,
		GetExports:   true,
		GetDebugInfo: true,
	})
	components := getAgentState(rawComponents)

	// TODO(rfratto): replace this with labels from config.
	o.client.SetAgentState(agentstate.NewAgentState(map[string]string{
		"hello": "world",
	}))
	o.client.SetComponents(components)

	if err := o.client.Send(ctx, o.args.RemoteEndpoint, "default"); err != nil {
		level.Error(o.log).Log("msg", "failed to send payload", "err", err)
	} else {
		level.Info(o.log).Log("msg", "sent state payload to remote server")
	}
}

func getAgentState(components []*component.Info) []agentstate.Component {
	res := []agentstate.Component{}

	for _, cInfo := range components {
		componentDetail := riverjson.GetComponentDetail(componentDetailInfo{
			Arguments: cInfo.Arguments,
			Exports:   cInfo.Exports,
			DebugInfo: cInfo.DebugInfo,
		})

		componentState := agentstate.Component{
			ID:       cInfo.ID.LocalID,
			ModuleID: cInfo.ID.ModuleID,
			Health: agentstate.Health{
				Health:     cInfo.Health.Health.String(),
				Message:    cInfo.Health.Message,
				UpdateTime: cInfo.Health.UpdateTime,
			},
			ComponentDetail: componentDetail,
		}

		res = append(res, componentState)
	}

	return res
}

type componentDetailInfo struct {
	Arguments any `river:"arguments,block"`
	Exports   any `river:"exports,block"`
	DebugInfo any `river:"debug_info,block"`
}

func getTopLevelComponentDetail(componentName string, parentId uint, idCounter *uint) agentstate.ComponentDetail {
	res := agentstate.ComponentDetail{
		ID:         *idCounter,
		ParentID:   parentId,
		Name:       componentName,
		Label:      "",
		RiverType:  "",
		RiverValue: []byte{},
	}
	*idCounter += 1
	return res
}

// Update implements service.Service.
func (o *Observer) Update(newConfig any) error {
	cfg, ok := newConfig.(Arguments)
	if !ok {
		return fmt.Errorf("invalid configuration passed to the %q service", ServiceName)
	}

	o.mtx.Lock()
	defer o.mtx.Unlock()

	o.args = cfg

	select {
	case o.configUpdate <- struct{}{}:
	default:
		// No-op; update is already scheduled.
	}

	return nil
}
