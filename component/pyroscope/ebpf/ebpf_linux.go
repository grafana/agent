//go:build linux

package ebpf

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	ebpfspy "github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/metrics"
	"github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy/sd"
	"github.com/oklog/run"
)

func init() {
	component.Register(component.Registration{
		Name: "pyroscope.ebpf",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

func New(o component.Options, args Arguments) (component.Component, error) {
	flowAppendable := pyroscope.NewFanout(args.ForwardTo, o.ID, o.Registerer)

	ms := metrics.NewMetrics(o.Registerer)

	tf, err := sd.NewTargetFinder(o.Logger, args.ContainerIDCacheSize, ms)
	if err != nil {
		return nil, fmt.Errorf("ebpf target finder create: %w", err)
	}

	session, err := ebpfspy.NewSession(
		o.Logger,
		tf,
		ms,
		args.SampleRate,
		ebpfspy.CacheOptions{
			PidCacheSize: args.PidCacheSize,
			ElfCacheSize: args.ElfCacheSize,
		},
		ebpfspy.ProfileOptions{
			CollectUser:   args.CollectUserProfile,
			CollectKernel: args.CollectKernelProfile,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("ebpf session create: %w", err)
	}
	res := &Component{
		options:      o,
		appendable:   flowAppendable,
		args:         args,
		targetFinder: tf,
		session:      session,
		argsUpdate:   make(chan Arguments),
	}
	res.updateTargetFinder()
	return res, nil
}

type Arguments struct {
	ForwardTo            []pyroscope.Appendable `river:"forward_to,attr"`
	Targets              []discovery.Target     `river:"targets,attr,optional"`
	DefaultTarget        discovery.Target       `river:"default_target,attr,optional"`
	TargetsOnly          bool                   `river:"targets_only,attr,optional"`
	CollectInterval      time.Duration          `river:"collect_interval,attr,optional"`
	SampleRate           int                    `river:"sample_rate,attr,optional"`
	PidCacheSize         int                    `river:"pid_cache_size,attr,optional"`
	ElfCacheSize         int                    `river:"elf_cache_size,attr,optional"`
	ContainerIDCacheSize int                    `river:"container_id_cache_size,attr,optional"`
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
		CollectInterval:      10 * time.Second,
		SampleRate:           100,
		PidCacheSize:         32,
		ContainerIDCacheSize: 64,
		ElfCacheSize:         128,
		TargetsOnly:          true,
		CollectUserProfile:   true,
		CollectKernelProfile: true,
	}
}

type Component struct {
	options      component.Options
	args         Arguments
	argsUpdate   chan Arguments
	appendable   *pyroscope.Fanout
	targetFinder *sd.TargetFinder
	session      *ebpfspy.Session
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
				c.updateTargetFinder()
				c.session.UpdateCacheOptions(ebpfspy.CacheOptions{
					PidCacheSize: c.args.PidCacheSize,
					ElfCacheSize: c.args.ElfCacheSize,
				})
				err := c.session.UpdateSampleRate(c.args.SampleRate)
				if err != nil {
					return nil
				}
				c.appendable.UpdateChildren(newArgs.ForwardTo)
				if c.args.CollectInterval != collectInterval {
					t.Reset(c.args.CollectInterval)
					collectInterval = c.args.CollectInterval
				}
			case <-t.C:
				err := c.reset()
				if err != nil {
					return err
				}
			}
		}
	}, func(error) {

	})
	return g.Run()
}

func (c *Component) updateTargetFinder() {
	c.targetFinder.SetTargets(sd.TargetsOptions{
		Targets:       c.args.Targets,
		DefaultTarget: c.args.DefaultTarget,
		TargetsOnly:   c.args.TargetsOnly,
	})
	c.targetFinder.ResizeContainerIDCache(c.args.ContainerIDCacheSize)
}

func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	c.argsUpdate <- newArgs
	return nil
}

func (c *Component) reset() error {
	args := c.args
	builders := ebpfspy.NewProfileBuilders(args.SampleRate)
	err := c.session.Reset(func(target *sd.Target, stack []string, value uint64, pid uint32) {
		labelsHash, labels := target.Labels()
		builder := builders.BuilderForTarget(labelsHash, labels)
		builder.AddSample(stack, value)
	})
	if err != nil {
		return fmt.Errorf("ebpf session reset %w", err)
	}
	for _, builder := range builders.Builders {
		var buf bytes.Buffer
		err := builder.Profile.Write(&buf)
		if err != nil {
			return fmt.Errorf("ebpf profile encode %w", err)
		}
		appender := c.appendable.Appender()
		samples := []*pyroscope.RawSample{{RawProfile: buf.Bytes()}}
		err = appender.Append(context.Background(), builder.Labels, samples)
		if err != nil {
			level.Error(c.options.Logger).Log("msg", "ebpf pprof write", "err", err)
			continue
		}
	}
	return nil
}
