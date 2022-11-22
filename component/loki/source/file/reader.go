package file

// Reader contains the set of methods the loki.source.file component uses.
type Reader interface {
	Stop()
	IsRunning() bool
	Path() string
	MarkPositionAndSize() error
}
