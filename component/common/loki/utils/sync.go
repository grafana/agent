package utils

import "sync"

// SyncSlice is a concurrent slice implementation.
type SyncSlice[T any] struct {
	list []T
	lock sync.RWMutex
}

func NewSyncSlice[T any]() *SyncSlice[T] {
	return &SyncSlice[T]{
		list: make([]T, 0),
	}
}

func (ss *SyncSlice[T]) Append(el T) {
	ss.lock.Lock()
	ss.list = append(ss.list, el)
	ss.lock.Unlock()
}

func (ss *SyncSlice[T]) Length() int {
	ss.lock.RLock()
	defer ss.lock.RUnlock()
	return len(ss.list)
}

// StartIterate returns the internal slice, after read-locking the internal lock. Once the iteration is finished,
// DoneIterate should be called to release the lock.
func (ss *SyncSlice[T]) StartIterate() []T {
	ss.lock.RLock()
	return ss.list
}

// DoneIterate releases the internal read-lock.
func (ss *SyncSlice[T]) DoneIterate() {
	ss.lock.RUnlock()
}
