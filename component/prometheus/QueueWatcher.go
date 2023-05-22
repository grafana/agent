package prometheus

import "github.com/prometheus/prometheus/tsdb/wlog"

type QueueWatcher interface {
	SetWriteTo(write wlog.WriteTo)
	Start()
	Stop()
}
