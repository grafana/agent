package util

import (
	"bytes"
	"sync"
)

// SyncBuffer wraps around a bytes.Buffer and makes it safe to use from
// multiple goroutines.
type SyncBuffer struct {
	mut sync.RWMutex
	buf bytes.Buffer
}

func (sb *SyncBuffer) Bytes() []byte {
	sb.mut.RLock()
	defer sb.mut.RUnlock()

	return sb.buf.Bytes()
}

func (sb *SyncBuffer) String() string {
	sb.mut.RLock()
	defer sb.mut.RUnlock()

	return sb.buf.String()
}

func (sb *SyncBuffer) Write(p []byte) (n int, err error) {
	sb.mut.Lock()
	defer sb.mut.Unlock()

	return sb.buf.Write(p)
}
