package java

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/pkg/flow/logging/level"
)

const (
	labelProcessID = "__process_pid__"
)

func init() {
	component.Register(component.Registration{
		Name: "pyroscope.java",
		Args: Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			err := profiler.Extract()
			if err != nil {
				return nil, fmt.Errorf("extract async profiler: %w", err)
			}
			a := args.(Arguments)
			flowAppendable := pyroscope.NewFanout(a.ForwardTo, opts.ID, opts.Registerer)
			c := &javaComponent{
				opts:      opts,
				forwardTo: flowAppendable,
			}
			c.updateTargets(a.Targets)
			return c, nil
		},
	})
}

type Arguments struct {
	Targets   []discovery.Target     `river:"targets,attr"`
	ForwardTo []pyroscope.Appendable `river:"forward_to,attr"`

	Interval time.Duration `river:"interval,attr,optional"`
}

type javaComponent struct {
	opts      component.Options
	forwardTo *pyroscope.Fanout

	mutex       sync.Mutex
	pid2process map[int]*profilingLoop
}

func (j *javaComponent) Run(ctx context.Context) error {
	defer func() {
		j.stop()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

func (j *javaComponent) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	j.forwardTo.UpdateChildren(newArgs.ForwardTo)
	a := args.(Arguments)
	j.updateTargets(a.Targets)
	return nil
}

// todo debug info

func (j *javaComponent) updateTargets(targets []discovery.Target) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	active := make(map[int]struct{})
	for _, target := range targets {
		fmt.Printf("target: %v\n", target)
		pid, err := strconv.Atoi(target[labelProcessID])
		if err != nil {
			_ = level.Error(j.opts.Logger).Log("msg", "invalid target", "target", fmt.Sprintf("%v", target), "err", err)
			continue
		}
		proc := j.pid2process[pid]
		if proc == nil {
			proc = newProcess(pid, target, j.opts.Logger, j.forwardTo)
		} else {
			proc.update(target)
		}
		active[pid] = struct{}{}
	}
	for pid := range j.pid2process {
		if _, ok := active[pid]; ok {
			continue
		}
		j.pid2process[pid].Close()
		delete(j.pid2process, pid)
	}
}

func (j *javaComponent) stop() {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	for _, proc := range j.pid2process {
		proc.Close()
		delete(j.pid2process, proc.pid)
	}
}
