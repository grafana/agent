package frontendcollector

import (
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/grafana/agent/pkg/loki"
)

type FrontendCollector struct {
	mut       sync.Mutex
	l         log.Logger
	instances map[string]*Instance
}

func New(c Config, loki *loki.Loki, l log.Logger) (*FrontendCollector, error) {
	frontendCollector := &FrontendCollector{
		instances: make(map[string]*Instance),
		l:         log.With(l, "component", "frontendcollector"),
	}

	if err := frontendCollector.ApplyConfig(loki, c); err != nil {
		return nil, err
	}

	return frontendCollector, nil
}

func (f *FrontendCollector) ApplyConfig(loki *loki.Loki, c Config) error {
	f.mut.Lock()
	defer f.mut.Unlock()

	newInstances := make(map[string]*Instance, len(c.Configs))

	for _, ic := range c.Configs {
		// If an old instance existed, update it and move it to the new map.
		if old, ok := f.instances[ic.Name]; ok {
			err := old.ApplyConfig(loki, ic)
			if err != nil {
				return err
			}

			newInstances[ic.Name] = old
			continue
		}

		inst, err := NewInstance(loki, ic, f.l)
		if err != nil {
			return fmt.Errorf("unable to apply config for %s: %w", ic.Name, err)
		}
		newInstances[ic.Name] = inst
	}

	// Any promtail in l.instances that isn't in newInstances has been removed
	// from the config. Stop them before replacing the map.
	for key, i := range f.instances {
		if _, exist := newInstances[key]; exist {
			continue
		}
		i.Stop()
	}
	f.instances = newInstances

	return nil
}

func (f *FrontendCollector) Stop() {
	f.mut.Lock()
	defer f.mut.Unlock()

	for _, i := range f.instances {
		i.Stop()
	}
}
