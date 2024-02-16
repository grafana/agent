package relabel

import (
	"context"
	"sync"

	"github.com/grafana/agent/component"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/discovery"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.relabel",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the discovery.relabel component.
type Arguments struct {
	// Targets contains the input 'targets' passed by a service discovery component.
	Targets []discovery.Target `river:"targets,attr"`

	// The relabelling rules to apply to each target's label set.
	RelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`
}

// Exports holds values which are exported by the discovery.relabel component.
type Exports struct {
	Output []discovery.Target `river:"output,attr"`
	Rules  flow_relabel.Rules `river:"rules,attr"`
}

// Component implements the discovery.relabel component.
type Component struct {
	opts component.Options

	mut sync.RWMutex
	rcs []*relabel.Config

	// todo: limit cache size. rotate out old entries, etc.
	cache     *lru.Cache[string, *discovery.Target]
	cacheSize int
}

const initialCacheSize = 500

var _ component.Component = (*Component)(nil)

// New creates a new discovery.relabel component.
func New(o component.Options, args Arguments) (*Component, error) {
	cache, err := lru.New[string, *discovery.Target](initialCacheSize)
	if err != nil {
		return nil, err
	}
	c := &Component{
		opts:      o,
		cache:     cache,
		cacheSize: initialCacheSize,
	}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	newArgs := args.(Arguments)

	// todo: if rules change, purge cache
	targets := make([]discovery.Target, 0, len(newArgs.Targets))

	relabelConfigs := flow_relabel.ComponentToPromRelabelConfigs(newArgs.RelabelConfigs)
	c.rcs = relabelConfigs

	c.EnsureCacheSize(len(newArgs.Targets))

	for _, t := range newArgs.Targets {
		key := t.GetHash()

		if newT, ok := c.cache.Get(key); ok {
			if newT != nil {
				targets = append(targets, *newT)
			}
			continue
		}
		lset := t.Labels()
		lset, keep := relabel.Process(lset, relabelConfigs...)
		if keep {
			targ := promLabelsToComponent(lset)
			targ.ResetHash()
			targets = append(targets, targ)
			c.cache.Add(key, &targ)
		} else {
			c.cache.Add(key, nil)
		}
	}

	c.opts.OnStateChange(Exports{
		Output: targets,
		Rules:  newArgs.RelabelConfigs,
	})

	return nil
}

func promLabelsToComponent(ls labels.Labels) discovery.Target {
	res := make(map[string]string, len(ls))
	for _, l := range ls {
		res[l.Name] = l.Value
	}

	return res
}

// EnsureCacheSize makes sure our lru cache is big enough for this target set.
// if it is too small it will not be useful
func (c *Component) EnsureCacheSize(size int) {
	// If it less than 1.25x the number of targets, increase to 1.5x?
	const minSizeFactor = 1.25
	const increaseSizeFactor = 1.5
	min := int(float64(size) * minSizeFactor)
	if c.cacheSize < min {
		newSize := int(float64(size) * increaseSizeFactor)
		c.cache.Resize(newSize)
		c.cacheSize = newSize
	}
	// TODO: possibly reduce size if we suddenly become way too large?
}
