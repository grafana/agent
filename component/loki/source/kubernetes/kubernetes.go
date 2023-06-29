// Package kubernetes implements the loki.source.kubernetes component.
package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	commonk8s "github.com/grafana/agent/component/common/kubernetes"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/positions"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/loki/source/kubernetes/kubetail"
	"k8s.io/client-go/kubernetes"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.kubernetes",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the loki.source.kubernetes
// component.
type Arguments struct {
	Targets   []discovery.Target  `river:"targets,attr"`
	ForwardTo []loki.LogsReceiver `river:"forward_to,attr"`

	// Client settings to connect to Kubernetes.
	Client commonk8s.ClientArguments `river:"client,block,optional"`
}

// DefaultArguments holds default settings for loki.source.kubernetes.
var DefaultArguments = Arguments{
	Client: commonk8s.ClientArguments{
		HTTPClientConfig: config.DefaultHTTPClientConfig,
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Component implements the loki.source.kubernetes component.
type Component struct {
	log       log.Logger
	opts      component.Options
	positions positions.Positions

	mut         sync.Mutex
	args        Arguments
	tailer      *kubetail.Manager
	lastOptions *kubetail.Options

	handler loki.LogsReceiver

	receiversMut sync.RWMutex
	receivers    []loki.LogsReceiver
}

var (
	_ component.Component      = (*Component)(nil)
	_ component.DebugComponent = (*Component)(nil)
)

// New creates a new loki.source.kubernetes component.
func New(o component.Options, args Arguments) (*Component, error) {
	err := os.MkdirAll(o.DataPath, 0750)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	positionsFile, err := positions.New(o.Logger, positions.Config{
		SyncPeriod:    10 * time.Second,
		PositionsFile: filepath.Join(o.DataPath, "positions.yml"),
	})

	if err != nil {
		return nil, err
	}

	c := &Component{
		log:       o.Logger,
		opts:      o,
		handler:   loki.NewLogsReceiver(),
		positions: positionsFile,
	}
	if err := c.Update(args); err != nil {
		return nil, err
	}
	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer c.positions.Stop()

	defer func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		// Guard for safety, but it's not possible for Run to be called without
		// c.tailer being initialized.
		if c.tailer != nil {
			c.tailer.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler.Chan():
			c.receiversMut.RLock()
			receivers := c.receivers
			c.receiversMut.RUnlock()

			for _, receiver := range receivers {
				receiver.Chan() <- entry
			}
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	// Update the receivers before anything else, just in case something fails.
	c.receiversMut.Lock()
	c.receivers = newArgs.ForwardTo
	c.receiversMut.Unlock()

	c.mut.Lock()
	defer c.mut.Unlock()

	managerOpts, err := c.getTailerOptions(newArgs)
	if err != nil {
		return err
	}

	switch {
	case c.tailer == nil:
		// First call to Update; build the tailer.
		c.tailer = kubetail.NewManager(c.log, managerOpts)

	case managerOpts != c.lastOptions:
		// Options changed; pass it to the tailer.
		//
		// This will never fail because it only fails if the context gets canceled.
		//
		// TODO(rfratto): should we have a generous update timeout to prevent this
		// from potentially hanging forever?
		_ = c.tailer.UpdateOptions(context.Background(), managerOpts)
		c.lastOptions = managerOpts

	default:
		// No-op: manager already exists and options didn't change.
	}

	// Convert input targets into targets to give to tailer.
	targets := make([]*kubetail.Target, 0, len(newArgs.Targets))

	for _, inTarget := range newArgs.Targets {
		lset := inTarget.Labels()
		processed, err := kubetail.PrepareLabels(lset, c.opts.ID)
		if err != nil {
			// TODO(rfratto): should this set the health of the component?
			level.Error(c.log).Log("msg", "failed to process input target", "target", lset.String(), "err", err)
			continue
		}
		targets = append(targets, kubetail.NewTarget(lset, processed))
	}

	// This will never fail because it only fails if the context gets canceled.
	//
	// TODO(rfratto): should we have a generous update timeout to prevent this
	// from potentially hanging forever?
	_ = c.tailer.SyncTargets(context.Background(), targets)

	c.args = newArgs
	return nil
}

// getTailerOptions gets tailer options from arguments. If args hasn't changed
// from the last call to getTailerOptions, c.lastOptions is returned.
// c.lastOptions must be updated by the caller.
//
// getTailerOptions must only be called when c.mut is held.
func (c *Component) getTailerOptions(args Arguments) (*kubetail.Options, error) {
	if reflect.DeepEqual(c.args.Client, args.Client) && c.lastOptions != nil {
		return c.lastOptions, nil
	}

	cfg, err := args.Client.BuildRESTConfig(c.log)
	if err != nil {
		return c.lastOptions, fmt.Errorf("building Kubernetes config: %w", err)
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return c.lastOptions, fmt.Errorf("building Kubernetes client: %w", err)
	}

	return &kubetail.Options{
		Client:    clientSet,
		Handler:   loki.NewEntryHandler(c.handler.Chan(), func() {}),
		Positions: c.positions,
	}, nil
}

// DebugInfo returns debug information for loki.source.kubernetes.
func (c *Component) DebugInfo() interface{} {
	var info DebugInfo

	for _, target := range c.tailer.Targets() {
		var lastError string
		if err := target.LastError(); err != nil {
			lastError = err.Error()
		}

		info.Targets = append(info.Targets, DebugInfoTarget{
			Labels:          target.Labels().Map(),
			DiscoveryLabels: target.DiscoveryLabels().Map(),
			LastError:       lastError,
			UpdateTime:      target.LastEntry().Local(),
		})
	}

	return info
}

// DebugInfo represents debug information for loki.source.kubernetes.
type DebugInfo struct {
	Targets []DebugInfoTarget `river:"target,block,optional"`
}

// DebugInfoTarget is debug information for an individual target being tailed
// for logs.
type DebugInfoTarget struct {
	Labels          map[string]string `river:"labels,attr,optional"`
	DiscoveryLabels map[string]string `river:"discovery_labels,attr,optional"`
	LastError       string            `river:"last_error,attr,optional"`
	UpdateTime      time.Time         `river:"update_time,attr,optional"`
}
