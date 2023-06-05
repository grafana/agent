package memory

import (
	"context"
	"time"

	"github.com/grafana/agent/component/prometheus"
)

type writer struct {
	keys       []uint64
	currentKey uint64
	to         prometheus.WriteTo
	db         *db
	ctx        context.Context
}

func (w *writer) start() {

	keys := w.db.getKeys("metrics")
	w.currentKey = keys[0]
	for {

		samples := w.db.getMetricRecords(w.currentKey)
		w.to.Append(samples)

		tmr := time.NewTimer(100 * time.Millisecond)
		select {
		case <-w.ctx.Done():
			return
		case <-tmr.C:
			continue
		}
	}
}
