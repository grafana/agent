package cluster

import (
	"context"
	"fmt"
	"github.com/grafana/agent/component"
	"sync"
)

func init() {
	component.Register(component.Registration{
		Name: "targetout",
		Args: TargetOutConfig{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return NewTargetOut(opts, args.(TargetOutConfig))
		},
	})
}

type TargetOut struct {
	mut sync.Mutex
	cfg TargetOutConfig
}

func (t *TargetOut) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (t *TargetOut) Update(args component.Arguments) error {
	t.mut.Lock()
	defer t.mut.Unlock()
	cfg := args.(TargetOutConfig)
	if cfg.Receiver == nil {
		return nil
	}
	cfg.Receiver.Register(func(kc []KeyConfig) {
		for _, target := range kc {
			fmt.Println(t.cfg.Name, "target name", target.Name, "target keyname", target.KeyName)
		}
	})
	return nil
}

type TargetOutConfig struct {
	Name     string          `hcl:"name,attr"`
	Receiver *TargetReceiver `hcl:"input"`
}

func NewTargetOut(_ component.Options, c TargetOutConfig) (*TargetOut, error) {
	t := &TargetOut{
		cfg: c,
	}
	return t, nil
}
