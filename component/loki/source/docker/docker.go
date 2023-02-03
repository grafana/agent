package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/docker/docker/client"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/positions"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/discovery"
	dt "github.com/grafana/agent/component/loki/source/docker/internal/dockertarget"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name: "loki.source.docker",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

const (
	dockerLabel                = model.MetaLabelPrefix + "docker_"
	dockerLabelContainerPrefix = dockerLabel + "container_"
	dockerLabelContainerID     = dockerLabelContainerPrefix + "id"
)

// Arguments holds values which are used to configure the loki.source.docker
// component.
type Arguments struct {
	Host         string              `river:"host,attr"`
	Targets      []discovery.Target  `river:"targets,attr"`
	ForwardTo    []loki.LogsReceiver `river:"forward_to,attr"`
	Labels       map[string]string   `river:"labels,attr,optional"`
	RelabelRules flow_relabel.Rules  `river:"relabel_rules,attr,optional"`
}

var (
	_ component.Component      = (*Component)(nil)
	_ component.DebugComponent = (*Component)(nil)
)

// Component implements the loki.source.file component.
type Component struct {
	opts    component.Options
	metrics *dt.Metrics

	mut           sync.RWMutex
	args          Arguments
	manager       *Manager
	lastOptions   *Options
	handler       loki.LogsReceiver
	receivers     []loki.LogsReceiver
	posFile       positions.Positions
	rcs           []*relabel.Config
	defaultLabels model.LabelSet
}

// New creates a new loki.source.file component.
func New(o component.Options, args Arguments) (*Component, error) {
	err := os.MkdirAll(o.DataPath, 0750)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	positionsFile, err := positions.New(o.Logger, positions.Config{
		SyncPeriod:        10 * time.Second,
		PositionsFile:     filepath.Join(o.DataPath, "positions.yml"),
		IgnoreInvalidYaml: false,
		ReadOnly:          false,
	})
	if err != nil {
		return nil, err
	}

	c := &Component{
		opts:    o,
		metrics: dt.NewMetrics(o.Registerer),

		handler:   make(loki.LogsReceiver),
		receivers: args.ForwardTo,
		posFile:   positionsFile,
	}

	// Call to Update() to start readers and set receivers once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer c.posFile.Stop()

	defer func() {
		c.mut.Lock()
		defer c.mut.Unlock()

		// Guard for safety, but it's not possible for Run to be called without
		// c.tailer being initialized.
		if c.manager != nil {
			c.manager.Stop()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry := <-c.handler:
			for _, receiver := range c.receivers {
				receiver <- entry
			}
		}
	}
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	c.mut.Lock()
	defer c.mut.Unlock()
	c.args = newArgs
	c.receivers = newArgs.ForwardTo

	managerOpts, err := c.getManagerOptions(newArgs)
	if err != nil {
		return err
	}

	switch {
	case c.manager == nil:
		// First call to Update; build the tailer.
		c.manager = NewManager(c.opts.Logger, managerOpts)
	case managerOpts != c.lastOptions:
		// Options changed; pass it to the tailer.
		// This will never fail because it only fails if the context gets canceled.
		_ = c.manager.UpdateOptions(context.Background(), managerOpts)
		c.lastOptions = managerOpts
	default:
		// No-op: manager already exists and options didn't change.
	}

	defaultLabels := make(model.LabelSet, len(newArgs.Labels))
	for k, v := range newArgs.Labels {
		defaultLabels[model.LabelName(k)] = model.LabelValue(v)
	}
	c.defaultLabels = defaultLabels

	if newArgs.RelabelRules != nil && len(newArgs.RelabelRules) > 0 {
		c.rcs = flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelRules)
	} else {
		c.rcs = []*relabel.Config{}
	}

	// Convert input targets into targets to give to tailer.
	targets := make([]*dt.Target, 0, len(newArgs.Targets))

	for _, target := range newArgs.Targets {
		containerID, ok := target[dockerLabelContainerID]
		if !ok {
			level.Debug(c.opts.Logger).Log("msg", "docker target did not include container ID label:"+dockerLabelContainerID)
			continue
		}

		var labels = make(model.LabelSet)
		for k, v := range target {
			labels[model.LabelName(k)] = model.LabelValue(v)
		}

		tgt, err := dt.NewTarget(
			c.metrics,
			log.With(c.opts.Logger, "target", fmt.Sprintf("docker/%s", containerID)),
			c.manager.opts.Handler,
			c.manager.opts.Positions,
			containerID,
			labels.Merge(c.defaultLabels),
			c.rcs,
			c.manager.opts.Client,
		)
		if err != nil {
			return err
		}
		targets = append(targets, tgt)

		// This will never fail because it only fails if the context gets canceled.
		_ = c.manager.SyncTargets(context.Background(), targets)
	}
	return nil
}

// getTailerOptions gets tailer options from arguments. If args hasn't changed
// from the last call to getTailerOptions, c.lastOptions is returned.
// c.lastOptions must be updated by the caller.
//
// getTailerOptions must only be called when c.mut is held.
func (c *Component) getManagerOptions(args Arguments) (*Options, error) {
	if reflect.DeepEqual(c.args.Host, args.Host) && c.lastOptions != nil {
		return c.lastOptions, nil
	}

	opts := []client.Opt{
		client.WithHost(c.args.Host),
		client.WithAPIVersionNegotiation(),
	}
	client, err := client.NewClientWithOpts(opts...)
	if err != nil {
		level.Error(c.opts.Logger).Log("msg", "could not create new Docker client", "err", err)
		return c.lastOptions, fmt.Errorf("failed to build docker client: %w", err)
	}

	return &Options{
		Client:    client,
		Handler:   loki.NewEntryHandler(c.handler, func() {}),
		Positions: c.posFile,
	}, nil
}

// DebugInfo returns information about the status of tailed targets.
func (c *Component) DebugInfo() interface{} {
	var res readerDebugInfo
	for _, tgt := range c.manager.Targets() {
		details := tgt.Details().(map[string]string)
		res.TargetsInfo = append(res.TargetsInfo, targetInfo{
			Labels:     tgt.Labels().String(),
			ID:         details["id"],
			LastError:  details["error"],
			IsRunning:  details["running"],
			ReadOffset: details["position"],
		})
	}
	return res
}

type readerDebugInfo struct {
	TargetsInfo []targetInfo `river:"targets_info,block"`
}

type targetInfo struct {
	ID         string `river:"id,attr"`
	LastError  string `river:"last_error,attr"`
	Labels     string `river:"labels,attr"`
	IsRunning  string `river:"is_running,attr"`
	ReadOffset string `river:"read_offset,attr"`
}
