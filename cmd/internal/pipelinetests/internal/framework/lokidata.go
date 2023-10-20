package framework

import (
	"strings"
	"sync"

	"github.com/grafana/loki/pkg/logproto"
)

type LokiData struct {
	mut        sync.Mutex
	lokiWrites []*logproto.PushRequest
}

func (r *LokiData) appendLokiWrite(req *logproto.PushRequest) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.lokiWrites = append(r.lokiWrites, req)
}

func (r *LokiData) WritesCount() int {
	r.mut.Lock()
	defer r.mut.Unlock()
	return len(r.lokiWrites)
}

func (r *LokiData) FindLineContaining(contents string) (*logproto.Entry, string) {
	r.mut.Lock()
	defer r.mut.Unlock()
	for i := len(r.lokiWrites) - 1; i >= 0; i-- {
		for _, stream := range r.lokiWrites[i].Streams {
			for _, entry := range stream.Entries {
				if strings.Contains(entry.Line, contents) {
					return &entry, stream.Labels
				}
			}
		}
	}
	return nil, ""
}
