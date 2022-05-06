package testcomponents

import (
	"context"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component"
)

func init() {
	component.Register(component.Registration{
		Name:    "testcomponents.passthrough",
		Config:  PassthroughConfig{},
		Exports: PassthroughExports{},

		Build: func(o component.Options, c component.Config) (component.Component, error) {
			return NewPassthrough(o, c.(PassthroughConfig))
		},
	})
}

// PassthroughConfig configures the testcomponents.passthrough component.
type PassthroughConfig struct {
	Input string `hcl:"input,attr"`
}

type PassthroughExports struct {
	Output string `hcl:"output,optional"`
}

// Passthrough implements the testcomponents.passthrough component, where it
// always emits its input as an output.
type Passthrough struct {
	opts component.Options
	log  log.Logger
}

func NewPassthrough(o component.Options, cfg PassthroughConfig) (*Passthrough, error) {
	t := &Passthrough{opts: o, log: o.Logger}
	if err := t.Update(cfg); err != nil {
		return nil, err
	}
	return t, nil
}

var (
	_ component.Component = (*Passthrough)(nil)
)

// Run implements Component.
func (t *Passthrough) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Run implements Component.
func (t *Passthrough) Update(newConfig component.Config) error {
	c := newConfig.(PassthroughConfig)

	level.Info(t.log).Log("msg", "passing through value", "value", c.Input)
	t.opts.OnStateChange(PassthroughExports{Output: c.Input})
	return nil
}
