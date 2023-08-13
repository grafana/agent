package simple

import (
	"context"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/pkg/flow/logging"
	"time"

	"github.com/grafana/agent/component/prometheus"
)

type writer struct {
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
	return &writer{
		parentId:    parent,
		keys:        make([]uint64, 0),
		currentKey:  0,
		to:          to,
		store:       store,
		bookmarkKey: name,
		l:           l,
	}
}

func (w *writer) Start(ctx context.Context) error {
	w.ctx = ctx
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
	for {
		val, signalFound := w.store.GetSignal(w.currentKey)
		if !signalFound {
			continue
		}
		switch v := val.(type) {
		case []prometheus.Sample:
			w.to.Append(v)
		case []prometheus.Metadata:
			w.to.AppendMetadata(v)
		case []prometheus.Exemplar:
			w.to.AppendExemplars(v)
		case []prometheus.FloatHistogram:
			w.to.AppendFloatHistograms(v)
		case []prometheus.Histogram:
			w.to.AppendHistograms(v)
		default:
			return fmt.Errorf("Unknown value %s ", v)
		}

		newKey, err := w.incrementKey()
		if err != nil {
			return err
		}

		timeOut := 100 * time.Millisecond
		// If there is a new key then dont wait long but we still need to check for the ctx being done.
		if newKey {
			level.Info(w.l).Log("key", w.currentKey)
			timeOut = 1 * time.Millisecond
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

// incrementKey returns true if key changed
func (w *writer) incrementKey() (bool, error) {
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
