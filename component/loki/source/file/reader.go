package file

// This code is copied from Promtail to accomodate the tailer and decompresser
// implementations as readers.

// reader contains the set of methods the loki.source.file component uses.
type reader interface {
	Stop()
	IsRunning() bool
	Path() string
	MarkPositionAndSize() error
}
