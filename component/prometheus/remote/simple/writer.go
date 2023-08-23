package simple

import (
	"context"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
	"sync"
	"time"

	"github.com/grafana/agent/component/prometheus"
)

type writer struct {
	mut         sync.RWMutex
	parentId    string
	keys        []uint64
	currentKey  uint64
	to          *QueueManager
	store       *dbstore
	ctx         context.Context
	bookmarkKey string
	l           *logging.Logger
}

func newWriter(parent string, to *QueueManager, store *dbstore, l *logging.Logger) *writer {

	name := fmt.Sprintf("metrics_write_to_%s_parent_%s", to.Name(), parent)
	w := &writer{
		parentId:    parent,
		keys:        make([]uint64, 0),
		currentKey:  0,
		to:          to,
		store:       store,
		bookmarkKey: name,
		l:           l,
	}
	v, found := w.store.GetBookmark(w.bookmarkKey)
	// If we dont have a bookmark then grab the oldest key.
	if !found {
		w.currentKey = w.store.GetOldestKey()
	} else {
		w.currentKey = v.Key
	}
	if w.currentKey == 0 {
		w.currentKey = 1
	}
	return w
}

func (w *writer) Start(ctx context.Context) error {
	w.mut.Lock()
	w.ctx = ctx
	w.mut.Unlock()

	newKey := false
	var err error
	success := true
	first := true
	for {
		timeOut := 10 * time.Second

		// If there is a new key then dont wait long but we still need to check for the ctx being done.
		// TODO this code is starting to get ugly, separate it out.
		if success {
			newKey, err = w.incrementKey()
			if err != nil {
				return err
			}
		}
		// If this is the first record we ALWAYS want to try sending.
		if newKey || first || !success {
			first = false
			level.Info(w.l).Log("msg", "looking for signal", "key", w.currentKey)
			// Eventually this will expire from the TTL.
			val, signalFound := w.store.GetSignal(w.currentKey)
			if signalFound {
				switch v := val.(type) {
				case []prometheus.Sample:
					success = w.to.Append(v)
				case []prometheus.Metadata:
					success = w.to.AppendMetadata(v)
				case []prometheus.Exemplar:
					success = w.to.AppendExemplars(v)
				case []prometheus.FloatHistogram:
					success = w.to.AppendFloatHistograms(v)
				case []prometheus.Histogram:
					success = w.to.AppendHistograms(v)
				default:
					return fmt.Errorf("Unknown value %s ", v)
				}
			} else {
				// No signal found so move on
				success = true
			}
		}

		level.Info(w.l).Log("msg", "sending success", "success", success)
		// If we were successful and have a newkey the quickly move on.
		if success && newKey {
			timeOut = 10 * time.Millisecond
		}

		tmr := time.NewTimer(timeOut)
		select {
		case <-w.ctx.Done():
			return nil
		case <-tmr.C:
			continue
		}
	}
}

func (w *writer) GetKey() uint64 {
	w.mut.RLock()
	defer w.mut.RUnlock()

	return w.currentKey
}

// incrementKey returns true if key changed
func (w *writer) incrementKey() (bool, error) {
	w.mut.Lock()
	defer w.mut.Unlock()

	prev := w.currentKey
	w.currentKey = w.store.GetNextKey(w.currentKey)
	// No need to update bookmark if nothing has changed.
	if prev == w.currentKey {
		return false, nil
	}
	err := w.store.WriteBookmark(w.bookmarkKey, &Bookmark{
		Key: w.currentKey,
	})
	return true, err
}
