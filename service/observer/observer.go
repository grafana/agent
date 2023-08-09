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
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/river/encoding/riveragentstate"
	"github.com/grafana/agent/service"
)

// ServiceName defines the name used for the Observer service.
const ServiceName = "observer"

type Arguments struct {
	RefreshFrequency time.Duration     `river:"refresh_frequency,attr,optional"`
	RemoteEndpoint   string            `river:"remote_endpoint,attr,optional"`
	Labels           map[string]string `river:"labels,attr,optional"`
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
	log     log.Logger
	agentID string

	mtx          sync.Mutex
	args         Arguments
	configUpdate chan struct{}

	client *ParquetClient
}

var _ service.Service = (*Observer)(nil)

// New returns a new, unstarted instance of the HTTP service.
func New(l log.Logger, agentID string) *Observer {
	return &Observer{
		log:          l,
		agentID:      agentID,
		configUpdate: make(chan struct{}, 1),
		args:         DefaultArguments,

		//TODO: Why not just set client to nil?
		//      We have not fully loaded the observer config yet. We can't send the any data anyway.
		client: NewParquetClient(
			riveragentstate.NewAgentState(nil),
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
	o.mtx.Lock()
	defer o.mtx.Unlock()

	if o.args.RemoteEndpoint == "" {
		// No server to send state to; skipping.
		return
	}

	level.Info(o.log).Log("msg", "sending state payload to remote server")

	rawComponents := component.GetAllComponents(host, component.InfoOptions{
		GetHealth:    true,
		GetArguments: true,
		GetExports:   true,
		GetDebugInfo: true,
	})
	components := getAgentState(rawComponents)

	// Copy the labels so that o.client doesn't reference the map inside the observer.
	//TODO: Would it be cleaner and safer if NewAgentState copies the map?
	//      Not sure what the conventions are in such cases.
	labelsCopy := make(map[string]string)
	for k, v := range o.args.Labels {
		labelsCopy[k] = v
	}
	o.client.SetAgentState(riveragentstate.NewAgentState(labelsCopy))
	o.client.SetComponents(components)

	if err := o.client.Send(ctx, o.args.RemoteEndpoint, o.agentID, "default"); err != nil {
		level.Error(o.log).Log("msg", "failed to send payload", "err", err)
	} else {
		level.Info(o.log).Log("msg", "sent state payload to remote server")
	}
}

func getAgentState(components []*component.Info) []riveragentstate.Component {
	res := []riveragentstate.Component{}

	for _, cInfo := range components {
		var (
			args      = riveragentstate.GetComponentDetail(cInfo.Arguments)
			exports   = riveragentstate.GetComponentDetail(cInfo.Exports)
			debugInfo = riveragentstate.GetComponentDetail(cInfo.DebugInfo)
		)

		componentState := riveragentstate.Component{
			ID:       cInfo.ID.LocalID,
			ModuleID: cInfo.ID.ModuleID,
			Health: riveragentstate.Health{
				Health:     cInfo.Health.Health.String(),
				Message:    cInfo.Health.Message,
				UpdateTime: cInfo.Health.UpdateTime,
			},
			Arguments: args,
			Exports:   exports,
			DebugInfo: debugInfo,
		}

		res = append(res, componentState)
	}

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
