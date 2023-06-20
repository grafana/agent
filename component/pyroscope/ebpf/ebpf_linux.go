//go:build linux

package ebpf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/pprof"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/sd"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/symtab"
	"github.com/oklog/run"
)

func init() {
	component.Register(component.Registration{
		Name: "pyroscope.ebpf",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			arguments := args.(Arguments)

			targetFinder, err := sd.NewTargetFinder(os.DirFS("/"), opts.Logger, targetsOptionFromArgs(arguments))
			if err != nil {
				return nil, fmt.Errorf("ebpf target finder create: %w", err)
			}

			session, err := ebpfspy.NewSession(
				opts.Logger,
				targetFinder,
				sessionOptionsFromArgs(arguments),
			)
			if err != nil {
				return nil, fmt.Errorf("ebpf session create: %w", err)
			}

			return New(opts, arguments, session, targetFinder)
		},
	})
}

func New(o component.Options, args Arguments, session ebpfspy.Session, targetFinder sd.TargetFinder) (component.Component, error) {
	flowAppendable := pyroscope.NewFanout(args.ForwardTo, o.ID, o.Registerer)

	metrics := newMetrics(o.Registerer)

	res := &Component{
		options:      o,
		metrics:      metrics,
		appendable:   flowAppendable,
		args:         args,
		targetFinder: targetFinder,
		session:      session,
		argsUpdate:   make(chan Arguments),
	}
	res.metrics.targetsActive.Set(float64(len(res.targetFinder.DebugInfo())))
	return res, nil
}

type Arguments struct {
	ForwardTo            []pyroscope.Appendable `river:"forward_to,attr"`
	Targets              []discovery.Target     `river:"targets,attr,optional"`
	CollectInterval      time.Duration          `river:"collect_interval,attr,optional"`
	SampleRate           int                    `river:"sample_rate,attr,optional"`
	PidCacheSize         int                    `river:"pid_cache_size,attr,optional"`
	BuildIDCacheSize     int                    `river:"build_id_cache_size,attr,optional"`
	SameFileCacheSize    int                    `river:"same_file_cache_size,attr,optional"`
	ContainerIDCacheSize int                    `river:"container_id_cache_size,attr,optional"`
	CacheRounds          int                    `river:"cache_rounds,attr,optional"`
	CollectUserProfile   bool                   `river:"collect_user_profile,attr,optional"`
	CollectKernelProfile bool                   `river:"collect_kernel_profile,attr,optional"`
}

func (rc *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*rc = defaultArguments()
	type config Arguments
	return f((*config)(rc))
}

func defaultArguments() Arguments {
	return Arguments{
		CollectInterval:      15 * time.Second,
		SampleRate:           97,
		PidCacheSize:         32,
		ContainerIDCacheSize: 1024,
		BuildIDCacheSize:     64,
		SameFileCacheSize:    8,
		CacheRounds:          3,
		CollectUserProfile:   true,
		CollectKernelProfile: true,
	}
}

type Component struct {
	options      component.Options
	args         Arguments
	argsUpdate   chan Arguments
	appendable   *pyroscope.Fanout
	targetFinder sd.TargetFinder
	session      ebpfspy.Session

	debugInfo     DebugInfo
	debugInfoLock sync.Mutex
	metrics       *metrics
}

func (c *Component) Run(ctx context.Context) error {
	err := c.session.Start()
	if err != nil {
		return fmt.Errorf("ebpf profiling session start: %w", err)
	}
	defer c.session.Stop()

	var g run.Group
	g.Add(func() error {
		collectInterval := c.args.CollectInterval
		t := time.NewTicker(collectInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case newArgs := <-c.argsUpdate:
				c.args = newArgs
				c.targetFinder.Update(targetsOptionFromArgs(c.args))
				c.metrics.targetsActive.Set(float64(len(c.targetFinder.DebugInfo())))
				err := c.session.Update(sessionOptionsFromArgs(c.args))
				if err != nil {
					return nil
				}
				c.appendable.UpdateChildren(newArgs.ForwardTo)
				if c.args.CollectInterval != collectInterval {
					t.Reset(c.args.CollectInterval)
					collectInterval = c.args.CollectInterval
				}
			case <-t.C:
				err := c.collectProfiles()
				if err != nil {
					c.metrics.profilingSessionsFailingTotal.Inc()
					return err
				}
				c.updateDebugInfo()
			}
		}
	}, func(error) {

	})
	return g.Run()
}

func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	c.argsUpdate <- newArgs
	return nil
}

func (c *Component) DebugInfo() interface{} {
	c.debugInfoLock.Lock()
	defer c.debugInfoLock.Unlock()
	return c.debugInfo
}

func (c *Component) collectProfiles() error {
	c.metrics.profilingSessionsTotal.Inc()
	level.Debug(c.options.Logger).Log("msg", "ebpf  collectProfiles")
	args := c.args
	builders := pprof.NewProfileBuilders(args.SampleRate)
	err := c.session.CollectProfiles(func(target *sd.Target, stack []string, value uint64, pid uint32) {
		labelsHash, labels := target.Labels()
		builder := builders.BuilderForTarget(labelsHash, labels)
		builder.AddSample(stack, value)
	})

	if err != nil {
		return fmt.Errorf("ebpf session collectProfiles %w", err)
	}
	level.Debug(c.options.Logger).Log("msg", "ebpf collectProfiles done", "profiles", len(builders.Builders))
	bytesSent := 0
	for _, builder := range builders.Builders {
		c.metrics.pprofsTotal.Inc()

		buf := bytes.NewBuffer(nil)
		_, err := builder.Write(buf)
		if err != nil {
			return fmt.Errorf("ebpf profile encode %w", err)
		}
		rawProfile := buf.Bytes()

		appender := c.appendable.Appender()
		bytesSent += len(rawProfile)
		samples := []*pyroscope.RawSample{{RawProfile: rawProfile}}
		err = appender.Append(context.Background(), builder.Labels, samples)
		if err != nil {
			level.Error(c.options.Logger).Log("msg", "ebpf pprof write", "err", err)
			continue
		}
	}
	level.Debug(c.options.Logger).Log("msg", "ebpf append done", "bytes_sent", bytesSent)
	return nil
}

type DebugInfo struct {
	Targets interface{} `river:"targets,attr,optional"`
	Session interface{} `river:"session,attr,optional"`
}

func (c *Component) updateDebugInfo() {
	c.debugInfoLock.Lock()
	defer c.debugInfoLock.Unlock()

	c.debugInfo = DebugInfo{
		Targets: c.targetFinder.DebugInfo(),
		Session: c.session.DebugInfo(),
	}
}

func targetsOptionFromArgs(args Arguments) sd.TargetsOptions {
	return sd.TargetsOptions{
		Targets:            args.Targets,
		DefaultTarget:      nil,
		TargetsOnly:        true,
		ContainerCacheSize: args.ContainerIDCacheSize,
	}
}

func cacheOptionsFromArgs(args Arguments) symtab.CacheOptions {
	return symtab.CacheOptions{
		PidCacheOptions: symtab.GCacheOptions{
			Size:       args.PidCacheSize,
			KeepRounds: args.CacheRounds,
		},
		BuildIDCacheOptions: symtab.GCacheOptions{
			Size:       args.BuildIDCacheSize,
			KeepRounds: args.CacheRounds,
		},
		SameFileCacheOptions: symtab.GCacheOptions{
			Size:       args.SameFileCacheSize,
			KeepRounds: args.CacheRounds,
		},
	}
}

func sessionOptionsFromArgs(args Arguments) ebpfspy.SessionOptions {
	return ebpfspy.SessionOptions{
		CollectUser:   args.CollectUserProfile,
		CollectKernel: args.CollectKernelProfile,
		SampleRate:    args.SampleRate,
		CacheOptions:  cacheOptionsFromArgs(args),
	}
}
