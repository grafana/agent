package metric

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestCounterExpiration(t *testing.T) {
	t.Parallel()
	cfg := &CounterConfig{
		Action:      "inc",
		Description: "HELP ME!!!!!",
		MaxIdle:     1 * time.Second,
	}

	cnt, err := NewCounters("test1", cfg)
	assert.Nil(t, err)

	// Create a label and increment the counter
	lbl1 := model.LabelSet{}
	lbl1["test"] = "i don't wanna make this a constant"
	cnt.With(lbl1).Inc()

	// Collect the metrics, should still find the metric in the map
	collect(cnt)
	assert.Contains(t, cnt.metrics, lbl1.Fingerprint())

	time.Sleep(1100 * time.Millisecond) // Wait just past our max idle of 1 sec

	//Add another counter with new label val
	lbl2 := model.LabelSet{}
	lbl2["test"] = "eat this linter"
	cnt.With(lbl2).Inc()

	// Collect the metrics, first counter should have expired and removed, second should still be present
	collect(cnt)
	assert.NotContains(t, cnt.metrics, lbl1.Fingerprint())
	assert.Contains(t, cnt.metrics, lbl2.Fingerprint())
}

func collect(c prometheus.Collector) {
	done := make(chan struct{})
	collector := make(chan prometheus.Metric)

	go func() {
		defer close(done)
		c.Collect(collector)
	}()

	for {
		select {
		case <-collector:
		case <-done:
			return
		}
	}
}
