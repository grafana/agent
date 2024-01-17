//go:build linux && (amd64 || arm64)

package dotnet

import (
	"context"
	"fmt"
	"sync"

	"github.com/grafana/agent/component"
	discdotnet "github.com/grafana/agent/component/discovery/dotnet"
	"github.com/grafana/agent/component/pyroscope"
	"github.com/grafana/agent/pkg/flow/logging/level"
)

func init() {
	component.Register(component.Registration{
		Name: "pyroscope.dotnet",
		Args: Arguments{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return &dotnetComponent{
				opts: opts,
				args: args.(Arguments),
			}, nil
		},
	})
}

type dotnetComponent struct {
	opts      component.Options
	args      Arguments
	forwardTo *pyroscope.Fanout

	mutex sync.Mutex
}

func (j *dotnetComponent) Run(ctx context.Context) error {
	defer func() {
		j.stop()
	}()
	<-ctx.Done()
	return nil
}

func (j *dotnetComponent) Update(args component.Arguments) error {
	newArgs := args.(Arguments)
	j.updateTargets(newArgs)
	return nil
}

func (j *dotnetComponent) updateTargets(args Arguments) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.args = args

	for _, target := range args.Targets {
		// get the socket
		socketPath := target[discdotnet.LabelDotnetDiagnosticSocket]
		fmt.Println("socketPath", socketPath)
	}
}

func (j *dotnetComponent) stop() {
	_ = level.Debug(j.opts.Logger).Log("msg", "stopping")
}
