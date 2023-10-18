package pipelinetests

import (
	"math"
	"sync"

	"github.com/prometheus/prometheus/prompb"
	"golang.org/x/exp/maps"
)

type promData struct {
	mut        sync.Mutex
	promWrites []*prompb.WriteRequest
}

func (r *promData) appendPromWrite(req *prompb.WriteRequest) {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.promWrites = append(r.promWrites, req)
}

func (r *promData) writesCount() int {
	r.mut.Lock()
	defer r.mut.Unlock()
	return len(r.promWrites)
}

func (r *promData) findLastSampleMatching(name string, labelsKV ...string) float64 {
	labelsMap := make(map[string]string)
	for i := 0; i < len(labelsKV); i += 2 {
		labelsMap[labelsKV[i]] = labelsKV[i+1]
	}
	labelsMap["__name__"] = name
	r.mut.Lock()
	defer r.mut.Unlock()
	// start from the end to find the most recent Timeseries
	for i := len(r.promWrites) - 1; i >= 0; i-- {
		for _, ts := range r.promWrites[i].Timeseries {
			// toMatch is a copy of labelsMap that we will remove labels from as we find matches
			toMatch := maps.Clone(labelsMap)
			for _, label := range ts.Labels {
				val, ok := toMatch[label.Name]
				if ok && val == label.Value {
					delete(toMatch, label.Name)
				}
			}
			foundMatch := len(toMatch) == 0
			if foundMatch && len(ts.Samples) > 0 {
				return ts.Samples[len(ts.Samples)-1].Value
			}
		}
	}
	return math.NaN()
}
