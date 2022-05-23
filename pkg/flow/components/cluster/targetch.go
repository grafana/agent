package cluster

import "sync"

type TargetReceiver struct {
	mut      sync.Mutex
	children []func([]KeyConfig)
}

func (t *TargetReceiver) Register(f func([]KeyConfig)) {
	// Need to track a unique name and be able to unregister, but lets pretend that works
	t.mut.Lock()
	defer t.mut.Unlock()
	t.children = append(t.children, f)
}

func (t *TargetReceiver) Send(kc []KeyConfig) {
	for _, f := range t.children {
		f(kc)
	}
}
