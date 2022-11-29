package scrape

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"

	"github.com/grafana/agent/component"
	component_config "github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pprof"
	"github.com/grafana/agent/component/prometheus/scrape"
)

const (
	pprofMemory     string = "memory"
	pprofBlock      string = "block"
	pprofGoroutine  string = "goroutine"
	pprofMutex      string = "mutex"
	pprofProcessCPU string = "process_cpu"
)

func init() {
	component.Register(component.Registration{
		Name: "pprof.scrape",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the pprof.scrape
// component.
type Arguments struct {
	Targets   []discovery.Target `river:"targets,attr"`
	ForwardTo []pprof.Appendable `river:"forward_to,attr"`

	// The job name to override the job label with.
	JobName string `river:"job_name,attr,optional"`
	// A set of query parameters with which the target is scraped.
	Params url.Values `river:"params,attr,optional"`
	// How frequently to scrape the targets of this scrape config.
	ScrapeInterval time.Duration `river:"scrape_interval,attr,optional"`
	// The timeout for scraping targets of this config.
	ScrapeTimeout time.Duration `river:"scrape_timeout,attr,optional"`
	// The URL scheme with which to fetch metrics from targets.
	Scheme string `river:"scheme,attr,optional"`

	// todo(ctovena): add support for limits.
	// // An uncompressed response body larger than this many bytes will cause the
	// // scrape to fail. 0 means no limit.
	// BodySizeLimit units.Base2Bytes `river:"body_size_limit,attr,optional"`
	// // More than this many targets after the target relabeling will cause the
	// // scrapes to fail.
	// TargetLimit uint `river:"target_limit,attr,optional"`
	// // More than this many labels post metric-relabeling will cause the scrape
	// // to fail.
	// LabelLimit uint `river:"label_limit,attr,optional"`
	// // More than this label name length post metric-relabeling will cause the
	// // scrape to fail.
	// LabelNameLengthLimit uint `river:"label_name_length_limit,attr,optional"`
	// // More than this label value length post metric-relabeling will cause the
	// // scrape to fail.
	// LabelValueLengthLimit uint `river:"label_value_length_limit,attr,optional"`

	HTTPClientConfig component_config.HTTPClientConfig `river:"http_client_config,block,optional"`

	ProfilingConfig *ProfilingConfig `river:"profiling_config,block,optional"`
}

type ProfilingConfig struct {
	PprofConfig PprofConfig `river:"pprof_config,attr,optional"`
	PprofPrefix string      `river:"path_prefix,attr,optional"`
}

type PprofConfig map[string]*PprofProfilingConfig

type PprofProfilingConfig struct {
	Enabled *bool  `river:"enabled,attr,optional"`
	Path    string `river:"path,attr,optional"`
	Delta   bool   `river:"delta,attr,optional"`
}

var DefaultArguments = NewDefaultArguments()

// NewDefaultArguments create the default settings for a scrape job.
func NewDefaultArguments() Arguments {
	return Arguments{
		Scheme:           "http",
		HTTPClientConfig: component_config.DefaultHTTPClientConfig,
		ScrapeInterval:   15 * time.Second,
		ScrapeTimeout:    15 * time.Second,
		ProfilingConfig: &ProfilingConfig{
			PprofConfig: PprofConfig{
				pprofMemory: &PprofProfilingConfig{
					Enabled: trueValue(),
					Path:    "/debug/pprof/allocs",
				},
				pprofBlock: &PprofProfilingConfig{
					Enabled: trueValue(),
					Path:    "/debug/pprof/block",
				},
				pprofGoroutine: &PprofProfilingConfig{
					Enabled: trueValue(),
					Path:    "/debug/pprof/goroutine",
				},
				pprofMutex: &PprofProfilingConfig{
					Enabled: trueValue(),
					Path:    "/debug/pprof/mutex",
				},
				pprofProcessCPU: &PprofProfilingConfig{
					Enabled: trueValue(),
					Delta:   true,
					Path:    "/debug/pprof/profile",
				},
			},
		},
	}
}

// UnmarshalRiver implements river.Unmarshaler.
func (arg *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*arg = NewDefaultArguments()

	type args Arguments
	if err := f((*args)(arg)); err != nil {
		return err
	}
	if arg.ProfilingConfig == nil || arg.ProfilingConfig.PprofConfig == nil {
		arg.ProfilingConfig = DefaultArguments.ProfilingConfig
	} else {
		for pt, pc := range DefaultArguments.ProfilingConfig.PprofConfig {
			if arg.ProfilingConfig.PprofConfig[pt] == nil {
				arg.ProfilingConfig.PprofConfig[pt] = pc
				continue
			}
			if arg.ProfilingConfig.PprofConfig[pt].Enabled == nil {
				arg.ProfilingConfig.PprofConfig[pt].Enabled = trueValue()
			}
			if arg.ProfilingConfig.PprofConfig[pt].Path == "" {
				arg.ProfilingConfig.PprofConfig[pt].Path = pc.Path
			}
		}
	}

	if arg.ScrapeTimeout > arg.ScrapeInterval {
		return fmt.Errorf("scrape timeout must be larger or equal to inverval")
	}
	if arg.ScrapeTimeout == 0 {
		arg.ScrapeTimeout = arg.ScrapeInterval
	}

	if cfg, ok := arg.ProfilingConfig.PprofConfig[pprofProcessCPU]; ok {
		if *cfg.Enabled && arg.ScrapeTimeout < time.Second*2 {
			return fmt.Errorf("%v scrape_timeout must be at least 2 seconds", pprofProcessCPU)
		}
	}

	return nil
}

// Component implements the pprof.scrape component.
type Component struct {
	opts component.Options

	reloadTargets chan struct{}

	mut        sync.RWMutex
	args       Arguments
	scraper    *Manager
	appendable *pprof.Fanout
}

var _ component.Component = (*Component)(nil)

// New creates a new pprof.scrape component.
func New(o component.Options, args Arguments) (*Component, error) {
	flowAppendable := pprof.NewFanout(args.ForwardTo, o.ID, o.Registerer)
	scraper := NewManager(flowAppendable, o.Logger)
	c := &Component{
		opts:          o,
		reloadTargets: make(chan struct{}, 1),
		scraper:       scraper,
		appendable:    flowAppendable,
	}

	// Call to Update() to set the receivers and targets once at the start.
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	defer c.scraper.Stop()

	targetSetsChan := make(chan map[string][]*targetgroup.Group)

	go func() {
		c.scraper.Run(targetSetsChan)
		level.Info(c.opts.Logger).Log("msg", "scrape manager stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.reloadTargets:
			c.mut.RLock()
			var (
				tgs     = c.args.Targets
				jobName = c.opts.ID
			)
			if c.args.JobName != "" {
				jobName = c.args.JobName
			}
			c.mut.RUnlock()
			promTargets := c.componentTargetsToProm(jobName, tgs)

			select {
			case targetSetsChan <- promTargets:
				level.Debug(c.opts.Logger).Log("msg", "passed new targets to scrape manager")
			case <-ctx.Done():
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

	c.appendable.UpdateChildren(newArgs.ForwardTo)

	err := c.scraper.ApplyConfig(newArgs)
	if err != nil {
		return fmt.Errorf("error applying scrape configs: %w", err)
	}
	level.Debug(c.opts.Logger).Log("msg", "scrape config was updated")

	select {
	case c.reloadTargets <- struct{}{}:
	default:
	}

	return nil
}

func (c *Component) componentTargetsToProm(jobName string, tgs []discovery.Target) map[string][]*targetgroup.Group {
	promGroup := &targetgroup.Group{Source: jobName}
	for _, tg := range tgs {
		promGroup.Targets = append(promGroup.Targets, convertLabelSet(tg))
	}

	return map[string][]*targetgroup.Group{jobName: {promGroup}}
}

func convertLabelSet(tg discovery.Target) model.LabelSet {
	lset := make(model.LabelSet, len(tg))
	for k, v := range tg {
		lset[model.LabelName(k)] = model.LabelValue(v)
	}
	return lset
}

// DebugInfo implements component.DebugComponent
func (c *Component) DebugInfo() interface{} {
	var res []scrape.TargetStatus

	for job, stt := range c.scraper.TargetsActive() {
		for _, st := range stt {
			var lastError string
			if st.LastError() != nil {
				lastError = st.LastError().Error()
			}
			if st != nil {
				// todo(ctovena): add more info
				res = append(res, scrape.TargetStatus{
					JobName:            job,
					URL:                st.URL().String(),
					Health:             string(st.Health()),
					Labels:             st.discoveredLabels.Map(),
					LastError:          lastError,
					LastScrape:         st.LastScrape(),
					LastScrapeDuration: st.LastScrapeDuration(),
				})
			}
		}
	}

	return scrape.ScraperStatus{TargetStatus: res}
}

func trueValue() *bool {
	a := true
	return &a
}
