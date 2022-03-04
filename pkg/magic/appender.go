package magic

import (
	"sync"
	"time"

	"github.com/prometheus/prometheus/pkg/timestamp"

	"github.com/prometheus/prometheus/pkg/exemplar"
	"github.com/prometheus/prometheus/pkg/labels"
)

type Appender struct {

	// values is a circular buffer of the last 100_000 values
	valueMutex sync.Mutex
	values     []Val
	valueIndex int
}

func newAppender() *Appender {
	return &Appender{values: make([]Val, 100_000)}
}

func (a *Appender) MetricValues() []Val {
	a.valueMutex.Lock()
	defer a.valueMutex.Unlock()
	cpy := make([]Val, len(a.values))
	copy(cpy, a.values)
	return cpy
}

func (a *Appender) Append(ref uint64, l labels.Labels, t int64, v float64) (uint64, error) {
	a.valueMutex.Lock()
	defer a.valueMutex.Unlock()
	if a.valueIndex == 100_000 {
		a.valueIndex = 0
	}
	a.values[a.valueIndex] = generateVal(l, t, v)
	a.valueIndex++
	return ref, nil
}

func (a *Appender) Commit() error {
	return nil
}

func (a *Appender) Rollback() error {
	return nil
}

func (a *Appender) AppendExemplar(_ uint64, l labels.Labels, e exemplar.Exemplar) (uint64, error) {
	return 0, nil
}

type Val struct {
	gathered  string
	timestamp string
	name      string
	labels    string
	value     float64
}

func (v *Val) Gathered() string {
	return v.gathered
}

func (v *Val) Timestamp() string {
	return v.timestamp
}

func (v *Val) Name() string {
	return v.name
}

func (v *Val) Labels() string {
	return v.labels
}

func (v *Val) Value() float64 {
	return v.value
}

func generateVal(l labels.Labels, t int64, v float64) Val {
	normalLabels := make(labels.Labels, 0)
	name := ""
	for _, item := range l {
		if item.Name == "__name__" {
			name = item.Value
			continue
		}
		normalLabels = append(normalLabels, item)
	}
	ts := timestamp.Time(t)

	return Val{
		gathered:  time.Now().Format("02/01/2006, 15:04:05"),
		timestamp: ts.Format("02/01/2006, 15:04:05"),
		name:      name,
		labels:    normalLabels.String(),
		value:     v,
	}
}
