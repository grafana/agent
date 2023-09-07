package qcache

import (
	"arena"
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/prometheus/prompb"
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
	l           log.Logger
}

func newWriter(parent string, to *QueueManager, store *dbstore, l log.Logger) *writer {
	name := fmt.Sprintf("metrics_write_to_%s_parent_%s", to.Name(), parent)
	w := &writer{
		parentId:    parent,
		keys:        make([]uint64, 0),
		currentKey:  0,
		to:          to,
		store:       store,
		bookmarkKey: name,
		l:           log.With(l, "name", name),
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

func (w *writer) Start(ctx context.Context) {
	w.mut.Lock()
	w.ctx = ctx
	w.mut.Unlock()

	// Always increment the key since we store the last written record.
	w.incrementKey()
	success := false
	bk := &Bookmark{}
	for {
		timeOut := 1 * time.Second
		// If we got a new key or the previous record did not enqueue then continue trying to send.
		if !w.isCurrentKey() || !success {
			w.incrementKey()
			valByte, _, signalFound := w.store.GetSignal(w.currentKey)
			if signalFound {
				var recoverableError bool
				success, recoverableError = w.send(valByte, ctx)
				// We need to succeed or hit an unrecoverable error to move on.
				if success || !recoverableError {
					// Write our bookmark of the last written record.
					bk.Key = w.currentKey
					err := w.store.WriteBookmark(w.bookmarkKey, bk)
					if err != nil {
						level.Error(w.l).Log("msg", "error writing bookmark", "err", err)
					}
				}
			}
		}

		// If we were successful and nothing is in the queue
		// If the queue is not full then give time for it to send.
		if success && !w.isCurrentKey() {
			timeOut = 10 * time.Millisecond
		}

		tmr := time.NewTimer(timeOut)
		select {
		case <-w.ctx.Done():
			return
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

func (w *writer) send(val []byte, ctx context.Context) (success bool, recoverableError bool) {
	recoverableError = true
	// Important that memory not be reused from samples or write request.
	mem := arena.NewArena()
	defer mem.Free()

	var err error
	samples, err := unmarshalSamples(bytes.NewBuffer(val), mem)
	for _, s := range samples {
		s.L, err = w.store.GetHashValue(s.Hash)
		if err != nil {
			level.Error(w.l).Log("msg", "unable to find hash", "err", err)
			return false, false
		}
	}
	if err != nil {
		level.Error(w.l).Log("msg", "error decoding samples", "err", err)
		return false, false
	}
	wr := arena.New[prompb.WriteRequest](mem)
	makeWriteRequest(samples, wr)
	success, err = w.to.Append(ctx, wr.Timeseries)
	if err != nil {
		// Let's check if it's an `out of order sample`. Yes this is some hand waving going on here.
		// TODO add metric for unrecoverable error
		if strings.Contains(err.Error(), "the sample has been rejected") {
			recoverableError = false
		}
		level.Error(w.l).Log("msg", "error sending samples", "err", err)
	}
	return success, recoverableError
}

// incrementKey returns true if key changed
func (w *writer) incrementKey() bool {
	w.mut.Lock()
	defer w.mut.Unlock()

	prev := w.currentKey
	w.currentKey = w.store.GetNextKey(w.currentKey)
	// No need to update bookmark if nothing has changed.
	return prev != w.currentKey
}

func (w *writer) isCurrentKey() bool {
	w.mut.Lock()
	defer w.mut.Unlock()

	nextKey := w.store.GetNextKey(w.currentKey)
	if nextKey != w.currentKey {
		level.Debug(w.l).Log("msg", "new key available", "current", w.currentKey, "next", nextKey)
	}
	return nextKey == w.currentKey
}

func makeWriteRequest(samples []*sample, wr *prompb.WriteRequest) {
	if len(samples) > len(wr.Timeseries) {
		wr.Timeseries = make([]prompb.TimeSeries, len(samples))
		for i := 0; i < len(samples); i++ {
			wr.Timeseries[i].Samples = make([]prompb.Sample, 1)
			wr.Timeseries[i].Samples[0] = prompb.Sample{}
			wr.Timeseries[i].Labels = make([]prompb.Label, 0)
		}
	}

	wr.Timeseries = wr.Timeseries[:len(samples)]
	for i, s := range samples {
		wr.Timeseries[i].Labels = bytesToLabels(s.L, wr.Timeseries[i].Labels)
		wr.Timeseries[i].Samples[0].Value = s.Value
		wr.Timeseries[i].Samples[0].Timestamp = s.TimeStamp
	}
}

func bytesToLabels(buf []byte, input []prompb.Label) []prompb.Label {
	// first byte is invalid.
	splitItems := bytes.Split(buf[1:], []byte{255})
	if input == nil || len(input) < len(splitItems)/2 {
		input = make([]prompb.Label, len(splitItems)/2)
	}
	index := 0
	for i := 0; i < len(splitItems); i = i + 2 {
		input[index].Name = string(splitItems[i])
		input[index].Value = string(splitItems[i+1])
		index++
	}
	return input
}
