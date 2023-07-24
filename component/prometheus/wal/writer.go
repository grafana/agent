package wal

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/grafana/agent/component/prometheus"
	"github.com/grafana/agent/component/prometheus/remote"
)

type writer struct {
	parentId   string
	keys       []uint64
	currentKey uint64
	to         remote.RemoteWrite
	bm         *bookmark
	metrics    *signaldb
	ctx        context.Context
}

func (w *writer) Start() error {
	name := fmt.Sprintf("metrics_write_to_%s_parent_%s", w.to.Name(), w.parentId)
	v, found, err := w.bm.getValueForKey(name)
	if err != nil {
		return err
	}
	// If we dont have a bookmark then grab the oldest key.
	if !found {
		keys, err := w.metrics.getKeys()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			w.currentKey = keys[0]
		}
	} else {
		// We do have a bookmark so read that.
		buf := bytes.NewBuffer(v)
		k, err := binary.ReadUvarint(buf)
		if err != nil {
			return err
		}
		w.currentKey = k
	}
	for {
		var samples []prometheus.Sample
		found, err := w.metrics.getRecordByUint(w.currentKey, samples)

		w.incrementKey()
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		w.to.Append(samples)

		tmr := time.NewTimer(100 * time.Millisecond)
		select {
		case <-w.ctx.Done():
			return nil
		case <-tmr.C:
			continue
		}
	}
}

func (w *writer) incrementKey() error {
	prev := w.currentKey
	w.currentKey = w.metrics.getNextKey(w.currentKey)
	// No need to update bookmark if nothing has changed.
	if prev == w.currentKey {
		return nil
	}
	name := fmt.Sprintf("metrics_write_to_%s_parent_%s", w.to.Name(), w.parentId)
	buf := make([]byte, 8)
	binary.PutUvarint(buf, w.currentKey)
	return w.bm.writeBookmark(name, buf)
}
  
