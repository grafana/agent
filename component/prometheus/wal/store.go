package wal

type Store interface {
	WriteBookmark(key string, value any) error
	GetBookmark(key string, into any) bool

	WriteSignal(table string, value any) (uint64, error)
	GetSignal(table string, key uint64, value any) bool

	WriteSignalCache(table string, key string, value any) error
	GetSignalCache(table string, key string, into any) bool

	RegisterTTLCallback(f func(table string, deletedIDs []uint64))
}
