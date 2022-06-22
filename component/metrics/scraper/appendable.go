package scraper

import (
	"context"
	"sync"

	"github.com/prometheus/prometheus/model/labels"

	"github.com/grafana/agent/component/metrics"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/storage"
)

type scrapeAppendable struct {
	mut       sync.Mutex
	buffer    map[int64][]metrics.FlowMetric
	receivers []*metrics.Receiver
}

func newScrapeAppendable(receiver []*metrics.Receiver) *scrapeAppendable {
	return &scrapeAppendable{
		buffer:    make(map[int64][]metrics.FlowMetric),
		receivers: receiver,
	}
}

func (s *scrapeAppendable) Append(ref storage.SeriesRef, l labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	// BIG TODO is that we probably want to move refid creation and caching into a shared cache between wal and scraper at some point.
	// 	in the below the refid is never cached by the scrape pool which makes it sad. Its likely this / refid caching should
	//  should create the refid cache
	s.mut.Lock()
	defer s.mut.Unlock()
	if len(s.receivers) == 0 {
		return 0, nil
	}
	_, found := s.buffer[t]
	if !found {
		set := make([]metrics.FlowMetric, 0)
		s.buffer[t] = set
	}
	s.buffer[t] = append(s.buffer[t], metrics.FlowMetric{
		Ref:    ref,
		Labels: l,
		Value:  v,
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
	s.mut.Unlock()
	for _, r := range s.receivers {
		for ts, metrics := range s.buffer {
			if r.Receive == nil {
				continue
			}
			r.Receive(ts, metrics)
		}
	}
	s.buffer = make(map[int64][]metrics.FlowMetric)
	return nil
}

func (s *scrapeAppendable) Rollback() error {
	s.buffer = make(map[int64][]metrics.FlowMetric)
	return nil
}

func (s *scrapeAppendable) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	//TODO implement me
	panic("implement me")
}

func (s *scrapeAppendable) Appender(ctx context.Context) storage.Appender {
	return s
}
