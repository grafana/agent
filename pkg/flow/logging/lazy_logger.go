package logging

import (
	"sync"

	"github.com/go-kit/log"
)

type lazyLogger struct {
	mut   sync.RWMutex
	inner log.Logger
}

func (ll *lazyLogger) UpdateInner(l log.Logger) {
	ll.mut.Lock()
	defer ll.mut.Unlock()
	ll.inner = l
}

func (ll *lazyLogger) Log(kvps ...interface{}) error {
	ll.mut.RLock()
	defer ll.mut.RUnlock()
	return ll.inner.Log(kvps...)
}
