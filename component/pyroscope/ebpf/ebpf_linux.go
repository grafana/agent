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
	"github.com/grafana/agent/component/pyroscope"
	ebpfspy "github.com/grafana/pyroscope/ebpf"
	"github.com/grafana/pyroscope/ebpf/pprof"
	"github.com/grafana/pyroscope/ebpf/sd"
	"github.com/grafana/pyroscope/ebpf/symtab"
	"github.com/grafana/pyroscope/ebpf/symtab/elf"
	"github.com/ianlancetaylor/demangle"
	"github.com/oklog/run"
)

func init() {
	component.Register(component.Registration{
		Name: "pyroscope.ebpf",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			arguments := args.(Arguments)
			return New(opts, arguments)
		},
	})
}

func New(opts component.Options, args Arguments) (component.Component, error) {
	targetFinder, err := sd.NewTargetFinder(os.DirFS("/"), opts.Logger, targetsOptionFromArgs(args))
	if err != nil {
		return nil, fmt.Errorf("ebpf target finder create: %w", err)
	}
	ms := newMetrics(opts.Registerer)

	session, err := ebpfspy.NewSession(
		opts.Logger,
		targetFinder,
		convertSessionOptions(args, ms),
	)
	if err != nil {
		return nil, fmt.Errorf("ebpf session create: %w", err)
	}

	flowAppendable := pyroscope.NewFanout(args.ForwardTo, opts.ID, opts.Registerer)

	res := &Component{
		options:      opts,
		metrics:      ms,
		appendable:   flowAppendable,
		args:         args,
		targetFinder: targetFinder,
		session:      session,
		argsUpdate:   make(chan Arguments),
	}
	res.metrics.targetsActive.Set(float64(len(res.targetFinder.DebugInfo())))
	return res, nil
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
		TargetsOnly:          true,
		Demangle:             "none",
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
				err := c.session.Update(convertSessionOptions(c.args, c.metrics))
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
		serviceName := builder.Labels.Get("service_name")
		c.metrics.pprofsTotal.WithLabelValues(serviceName).Inc()
		c.metrics.pprofSamplesTotal.WithLabelValues(serviceName).Add(float64(len(builder.Profile.Sample)))

		buf := bytes.NewBuffer(nil)
		_, err := builder.Write(buf)
		if err != nil {
			return fmt.Errorf("ebpf profile encode %w", err)
		}
		rawProfile := buf.Bytes()

		appender := c.appendable.Appender()
		bytesSent += len(rawProfile)
		c.metrics.pprofBytesTotal.WithLabelValues(serviceName).Add(float64(len(rawProfile)))

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
	targets := make([]sd.DiscoveryTarget, 0, len(args.Targets))
	for _, t := range args.Targets {
		targets = append(targets, sd.DiscoveryTarget(t))
	}
	return sd.TargetsOptions{
		Targets:            targets,
		DefaultTarget:      sd.DiscoveryTarget(args.DefaultTarget),
		TargetsOnly:        args.TargetsOnly,
		ContainerCacheSize: args.ContainerIDCacheSize,
	}
}

func convertSessionOptions(args Arguments, ms *metrics) ebpfspy.SessionOptions {
	return ebpfspy.SessionOptions{
		CollectUser:   args.CollectUserProfile,
		CollectKernel: args.CollectKernelProfile,
		SampleRate:    args.SampleRate,
		CacheOptions: symtab.CacheOptions{
			SymbolOptions: symtab.SymbolOptions{
				GoTableFallback: false,
				DemangleOptions: convertDemangleOptions(args.Demangle),
			},
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
			Metrics: ms.symtabMetrics,
		},
	}
}

func convertDemangleOptions(o string) []demangle.Option {
	switch o {
	case "none":
		return elf.DemangleNone
	case "simplified":
		return elf.DemangleSimplified
	case "templates":
		return elf.DemangleTemplates
	case "full":
		return elf.DemangleFull
	default:
		return elf.DemangleNone
	}
}
