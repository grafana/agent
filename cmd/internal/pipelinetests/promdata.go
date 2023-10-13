package pipelinetests

import (
	"math"
	"sync"

	"github.com/prometheus/prometheus/prompb"
	"golang.org/x/exp/slices"
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

func (r *promData) getPromWrites() []*prompb.WriteRequest {
	r.mut.Lock()
	defer r.mut.Unlock()
	return slices.Clone(r.promWrites)
}

func (r *promData) sampleValueForSeries(name string) float64 {
	r.mut.Lock()
	defer r.mut.Unlock()
	// start from the end to find the most recent Timeseries
	for i := len(r.promWrites) - 1; i >= 0; i-- {
		for _, ts := range r.promWrites[i].Timeseries {
			if ts.Labels[0].Name == "__name__" && ts.Labels[0].Value == name {
				return ts.Samples[len(ts.Samples)-1].Value
			}
		}
	}
	return math.NaN()
}
