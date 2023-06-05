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
	ebpfspy2 "github.com/grafana/agent/component/pyroscope/ebpf/ebpfspy"
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
	tf := sd.NewTargetFinder(o.Logger)

	session, err := ebpfspy2.NewSession(
		o.Logger,
		tf,
		uint32(args.SampleRate),
		args.PidCacheSize,
		args.ElfCacheSize,
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
	ForwardTo       []pyroscope.Appendable `river:"forward_to,attr"`
	Targets         []discovery.Target     `river:"targets,attr,optional"`
	DefaultTarget   discovery.Target       `river:"default_target,attr,optional"`
	KubernetesNode  string                 `river:"kubernetes_node,attr,optional"`
	TargetsOnly     bool                   `river:"targets_only,attr,optional"`
	ServiceName     string                 `river:"service_name,attr,optional"`
	CollectInterval time.Duration          `river:"collect_interval,attr,optional"`
	SampleRate      int                    `river:"sample_rate,attr,optional"`
	PidCacheSize    int                    `river:"pid_cache_size,attr,optional"`
	ElfCacheSize    int                    `river:"elf_cache_size,attr,optional"`
}

func (rc *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*rc = defaultArguments()
	type config Arguments
	return f((*config)(rc))
}

func defaultArguments() Arguments {
	return Arguments{
		CollectInterval: 10 * time.Second,
		SampleRate:      100,
		PidCacheSize:    64,
		ElfCacheSize:    128,
		TargetsOnly:     false,
	}
}

type Component struct {
	options      component.Options
	args         Arguments
	argsUpdate   chan Arguments
	appendable   *pyroscope.Fanout
	targetFinder *sd.TargetFinder
	session      *ebpfspy2.Session
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
				level.Debug(c.options.Logger).Log("msg", "args update ")
				c.args = newArgs
				c.updateTargetFinder()
				c.appendable.UpdateChildren(newArgs.ForwardTo)
				if c.args.CollectInterval != collectInterval {
					level.Debug(c.options.Logger).Log("msg", "reset timer to ", c.args.CollectInterval)
					t.Reset(c.args.CollectInterval)
					collectInterval = c.args.CollectInterval
				}
				level.Debug(c.options.Logger).Log("msg", "args update done")
			case <-t.C:
				level.Debug(c.options.Logger).Log("msg", "reset")
				err := c.reset()
				level.Debug(c.options.Logger).Log("msg", "reset done")
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
	c.targetFinder.SetTargets(sd.Options{
		Targets:       c.args.Targets,
		DefaultTarget: c.args.DefaultTarget,
		TargetsOnly:   c.args.TargetsOnly,
	})
}

func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	c.argsUpdate <- newArgs
	return nil
}

func (c *Component) reset() error {
	args := c.args
	builders := ebpfspy2.NewProfileBuilders(args.SampleRate)
	cnt := 0
	err := c.session.Reset(func(target *sd.Target, stack []string, value uint64, pid uint32) error {
		cnt++
		labelsHash, labels := target.Labels()
		builder := builders.BuilderForTarget(labelsHash, labels)
		builder.AddSample(stack, value)
		return nil
	})
	level.Debug(c.options.Logger).Log("msg", "ebpf session reset done, building pprofs...", "cnt", cnt)
	if err != nil {
		return fmt.Errorf("ebpf session reset %w", err)
	}
	for _, builder := range builders.Builders {
		level.Debug(c.options.Logger).Log(
			"msg", "ppof building",
			"target", builder.Labels.String(),
		)
		var buf bytes.Buffer
		err := builder.Profile.Write(&buf)
		if err != nil {
			return fmt.Errorf("ebpf profile encode %w", err)
		}
		appender := c.appendable.Appender()
		samples := []*pyroscope.RawSample{{RawProfile: buf.Bytes()}}
		level.Debug(c.options.Logger).Log(
			"msg", "ppof append",
			"target", builder.Labels.String(),
			"pprof", buf.Len(),
			"samples", len(builder.Profile.Sample),
		)
		err = appender.Append(context.Background(), builder.Labels, samples)
		level.Debug(c.options.Logger).Log(
			"msg", "ppof appended",
			"target", builder.Labels.String(),
			"res", err,
		)
		if err != nil {
			return fmt.Errorf("ebpf profile write %w", err)
		}
	}
	return nil
}
