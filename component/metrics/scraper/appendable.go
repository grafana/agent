package scraper

import (
	"context"
	"sync"

	"github.com/prometheus/prometheus/model/value"

	"github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/storage"
)

type scrapeAppendable struct {
	mut sync.Mutex
	// Though mostly a map of 1 item, this allows it to work if more than one TS gets added
	buffer    map[int64][]*metrics.FlowMetric
	receivers []*metrics.Receiver
}

func newScrapeAppendable(receiver []*metrics.Receiver) *scrapeAppendable {
	return &scrapeAppendable{
		buffer:    make(map[int64][]*metrics.FlowMetric),
		receivers: receiver,
	}
}

func (s *scrapeAppendable) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	if len(s.receivers) == 0 {
		return 0, nil
	}
	_, found := s.buffer[t]
	if !found {
		set := make([]*metrics.FlowMetric, 0)
		s.buffer[t] = set
	}
	// If ref is 0 then lets grab a global id
	if ref == 0 {
		ref = storage.SeriesRef(metrics.GlobalRefMapping.GetOrAddGlobalRefID(l))
	}
	// If it is stale then we can remove it
	if value.IsStaleNaN(v) {
		metrics.GlobalRefMapping.AddStaleMarker(uint64(ref), l)
	} else {
		metrics.GlobalRefMapping.RemoveStaleMarker(uint64(ref))
	}
	s.buffer[t] = append(s.buffer[t], &metrics.FlowMetric{
		GlobalRefID: uint64(ref),
		Labels:      l,
		Value:       v,
	})
	return ref, nil
}

func (s *scrapeAppendable) set(receiver []*metrics.Receiver) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.receivers = receiver
}

func (s *scrapeAppendable) Commit() error {
	s.mut.Lock()
	defer s.mut.Unlock()
	for _, r := range s.receivers {
		for ts, metrics := range s.buffer {
			if r.Receive == nil {
				continue
			}
			r.Receive(ts, metrics)
		}
	}
	s.buffer = make(map[int64][]*metrics.FlowMetric)
	return nil
}

func (s *scrapeAppendable) Rollback() error {
	s.buffer = make(map[int64][]*metrics.FlowMetric)
	return nil
}

func (s *scrapeAppendable) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	//TODO implement me
	panic("implement me")
}

func (s *scrapeAppendable) Appender(ctx context.Context) storage.Appender {
	return s
}
