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
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/service"
)

// ServiceName defines the name used for the Observer service.
const ServiceName = "observer"

type Arguments struct {
	RefreshFrequency time.Duration           `river:"refresh_frequency,attr,optional"`
	RemoteEndpoint   string                  `river:"remote_endpoint,attr,optional"`
	HTTPClientConfig config.HTTPClientConfig `river:",squash"`
	Headers          map[string]string       `river:"headers,attr,optional"`
	Labels           map[string]string       `river:"labels,attr,optional"`
}

var _ river.Defaulter = (*Arguments)(nil)

// DefaultArguments holds default values for Arguments.
var DefaultArguments = Arguments{
	RefreshFrequency: time.Minute,
	HTTPClientConfig: config.DefaultHTTPClientConfig,
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

	stateWriter agentStateWriter
}

var _ service.Service = (*Observer)(nil)

// New returns a new, unstarted instance of the HTTP service.
func New(l log.Logger, agentID string) *Observer {
	return &Observer{
		log:          l,
		agentID:      agentID,
		configUpdate: make(chan struct{}, 1),
		args:         DefaultArguments,
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
	o.stateWriter.SetContext(ctx)
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

	if o.stateWriter == nil {
		level.Error(o.log).Log("msg", "not sending agent state", "err", "no writer has been initialized")
		return
	}

	components := component.GetAllComponents(host, component.InfoOptions{
		GetHealth:    true,
		GetArguments: true,
		GetExports:   true,
		GetDebugInfo: true,
	})
	parquetRows := convertToParquetRows(components)

	parquet, err := createParquet(o.args.Labels, parquetRows)
	if err != nil {
		level.Error(o.log).Log("msg", "failed to create an agent state parquet file", "err", err)
		return
	}

	level.Info(o.log).Log("msg", "sending state payload to remote server")

	if _, err := o.stateWriter.Write(parquet); err != nil {
		level.Error(o.log).Log("msg", "failed to send payload", "err", err)
	} else {
		level.Info(o.log).Log("msg", "sent state payload to remote server")
	}
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

	// Copy the labels so that o.stateWriter doesn't reference the map inside the observer.
	labelsCopy := make(map[string]string)
	for k, v := range o.args.Labels {
		labelsCopy[k] = v
	}

	var err error
	o.stateWriter, err = newHttpAgentStateWriter(o.args.HTTPClientConfig, o.agentID, o.args.RemoteEndpoint, o.args.Headers)
	if err != nil {
		return fmt.Errorf("failed to create an HTTP agent state writer: %w", err)
	}

	select {
	case o.configUpdate <- struct{}{}:
	default:
		// No-op; update is already scheduled.
	}

	return nil
}
