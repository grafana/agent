package prometheus

type QueueWatcher interface {
	SetWriteTo(write WriteTo)
	Start()
	Stop()
}
