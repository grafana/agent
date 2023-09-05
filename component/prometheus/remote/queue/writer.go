package queue

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

	var err error
	newKey := w.incrementKey()
	success := true
	samples := make([]*sample, 0)
	bk := &Bookmark{}
	for {
		recoverableError := true
		timeOut := 1 * time.Second
		// If we got a new key or the previous record did not enqueue then continue trying to send.
		// TODO this is getting ugly
		if newKey || !success {
			level.Info(w.l).Log("msg", "looking for signal", "key", w.currentKey)
			valByte, _, signalFound := w.store.GetSignal(w.currentKey)
			if signalFound {
				mem := arena.NewArena()
				samples, err = unmarshalSamples(bytes.NewBuffer(valByte), mem)
				if err != nil {
					level.Error(w.l).Log("msg", "error decoding samples", "err", err)
					mem.Free()
				} else {

					wr := arena.New[prompb.WriteRequest](mem)
					makeWriteRequest(samples, wr)
					success, err = w.to.Append(ctx, wr.Timeseries)
					mem.Free()
					if err != nil {
						// Let's check if it's an `out of order sample`. Yes this is some hand waving going on here.
						// TODO add metric for unrecoverable error
						if strings.Contains(err.Error(), "the sample has been rejected") {
							recoverableError = false
						}
						level.Error(w.l).Log("msg", "error sending samples", "err", err)
					}

					// We need to succeed or hit an unrecoverable error to move on.
					if success || !recoverableError {
						// Write our bookmark of the last written record.
						bk.Key = w.currentKey
						err = w.store.WriteBookmark(w.bookmarkKey, bk)
						if err != nil {
							level.Error(w.l).Log("msg", "error writing bookmark", "err", err)
						}
					}
				}
			}
		}

		if success || !recoverableError {
			newKey = w.incrementKey()
		}

		// If we were successful and have a newkey the quickly move on.
		// If the queue is not full then give time for it to send.
		if success && newKey {
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

// incrementKey r github.com/zhuangsirui/binpackereturns true if key changed
func (w *writer) incrementKey() bool {
	w.mut.Lock()
	defer w.mut.Unlock()

	prev := w.currentKey
	w.currentKey = w.store.GetNextKey(w.currentKey)
	// No need to update bookmark if nothing has changed.
	return prev != w.currentKey
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
		wr.Timeseries[i].Labels = stringToLabelsProto(s.L, wr.Timeseries[i].Labels)
		wr.Timeseries[i].Samples[0].Value = s.Value
		wr.Timeseries[i].Samples[0].Timestamp = s.TimeStamp
	}
}

// labelsToLabelsProto transforms labels into prompb labels. The buffer slice
// will be used to avoid allocations if it is big enough to store the labels.
func stringToLabelsProto(lbls []string, input []prompb.Label) []prompb.Label {
	if input == nil || len(input) < len(lbls)/2 {
		input = make([]prompb.Label, len(lbls)/2)
	}
	index := 0
	for i := 0; i < len(lbls)/2; i++ {
		input[index] = prompb.Label{
			Name:  lbls[i+i],
			Value: lbls[i+i+1],
		}
		index++
	}

	return input[:len(lbls)/2]
}
