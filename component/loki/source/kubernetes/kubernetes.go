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
	"github.com/grafana/agent/component/common/loki"
	"github.com/grafana/agent/component/common/loki/positions"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/loki/source/kubernetes/kubetail"
	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/river"
	promconfig "github.com/prometheus/common/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
	Client ClientArguments `river:"client,block,optional"`
}

var _ river.Unmarshaler = (*Arguments)(nil)

// DefaultArguments holds default settings for loki.source.kubernetes.
var DefaultArguments = Arguments{
	Client: ClientArguments{
		HTTPClientConfig: config.DefaultHTTPClientConfig,
	},
}

// UnmarshalRiver implements river.Unmarshaler and applies defaults.
func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	return f((*arguments)(args))
}

// ClientArguments controls how loki.source.kubernetes connects to Kubernetes.
type ClientArguments struct {
	APIServer        config.URL              `river:"api_server,attr,optional"`
	KubeConfig       string                  `river:"kubeconfig_file,attr,optional"`
	HTTPClientConfig config.HTTPClientConfig `river:",squash"`
}

// UnmarshalRiver unmarshals ClientArguments and performs validations.
func (args *ClientArguments) UnmarshalRiver(f func(interface{}) error) error {
	type arguments ClientArguments
	if err := f((*arguments)(args)); err != nil {
		return err
	}

	if args.APIServer.URL != nil && args.KubeConfig != "" {
		return fmt.Errorf("only one of api_server and kubeconfig_file can be set")
	}
	if args.KubeConfig != "" && !reflect.DeepEqual(args.HTTPClientConfig, config.DefaultHTTPClientConfig) {
		return fmt.Errorf("custom HTTP client configuration is not allowed when kubeconfig_file is set")
	}
	if args.APIServer.URL == nil && !reflect.DeepEqual(args.HTTPClientConfig, config.DefaultHTTPClientConfig) {
		return fmt.Errorf("api_server must be set when custom HTTP client configuration is provided")
	}

	// We must explicitly Validate because HTTPClientConfig is squashed and it won't run otherwise
	return args.HTTPClientConfig.Validate()
}

// BuildRESTConfig converts ClientArguments to a Kubernetes REST config.
func (args *ClientArguments) BuildRESTConfig(l log.Logger) (*rest.Config, error) {
	var (
		cfg *rest.Config
		err error
	)

	switch {
	case args.KubeConfig != "":
		cfg, err = clientcmd.BuildConfigFromFlags("", args.KubeConfig)
		if err != nil {
			return nil, err
		}

	case args.APIServer.URL == nil:
		// Use in-cluster config.
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		level.Info(l).Log("msg", "Using pod service account via in-cluster config")

	default:
		rt, err := promconfig.NewRoundTripperFromConfig(*args.HTTPClientConfig.Convert(), "loki.source.kubernetes")
		if err != nil {
			return nil, err
		}
		cfg = &rest.Config{
			Host:      args.APIServer.String(),
			Transport: rt,
		}
	}

	cfg.UserAgent = fmt.Sprintf("GrafanaAgent/%s", build.Version)
	cfg.ContentType = "application/vnd.kubernetes.protobuf"

	return cfg, nil
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
		handler:   make(loki.LogsReceiver),
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
		case entry := <-c.handler:
			c.receiversMut.RLock()
			receivers := c.receivers
			c.receiversMut.RUnlock()

			for _, receiver := range receivers {
				receiver <- entry
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
		Handler:   loki.NewEntryHandler(c.handler, func() {}),
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

// DebugInfo represents debug information for loki.source.kubernets.
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
