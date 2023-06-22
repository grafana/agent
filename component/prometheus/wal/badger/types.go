package badger

import "github.com/grafana/agent/component/prometheus"

type seqSample struct {
	seq uint64
	prometheus.Sample
}

func (s *seqSample) SetSeq(seq uint64) {
	s.seq = seq
}
