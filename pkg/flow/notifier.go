package flow

import (
	"sync"
)

type NewFlow struct {
	F *Flow
}

type ClosedFlow struct {
	F *Flow
}

type Notifier struct {
	mut   sync.RWMutex
	Ch    chan interface{}
	Flows map[*Flow]struct{}
}

func (n *Notifier) Run() {
	for {
		select {
		case m := <-n.Ch:
			n.handleMessage(m)
		}
	}
}

func (n *Notifier) ComponentInfos() []*ComponentInfo {
	n.mut.RLock()
	defer n.mut.RUnlock()

	infos := make([]*ComponentInfo, 0)
	for k := range n.Flows {
		infos = append(infos, k.ComponentInfos()...)
	}
	return infos
}

func (n *Notifier) handleMessage(m interface{}) {
	n.mut.Lock()
	defer n.mut.Unlock()

	switch tt := m.(type) {
	case *NewFlow:
		n.Flows[tt.F] = struct{}{}
	case *ClosedFlow:
		delete(n.Flows, tt.F)
	}
}
