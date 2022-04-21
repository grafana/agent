package exchange

import (
	"github.com/iancoleman/orderedmap"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"time"
)

type Metric struct {
	name  string
	value float64
	ts    time.Time
	// Ordered map might not be needed like it is for logs
	labels []Label
}

type Label struct {
	Key   string
	Value string
}

func NewMetric(name string, value float64, ts time.Time, labels []Label) Metric {
	m := Metric{
		name:   name,
		value:  value,
		ts:     ts,
		labels: labels,
	}
	if _, found := m.FindLabel("__name__"); !found {
		m.labels = append(m.labels, Label{
			Key:   "__name__",
			Value: name,
		})
	}
	return m
}

func NewMetricFromPromMetric(ts int64, value float64, labels labels.Labels) Metric {
	name := ""
	newLabels := make([]Label, 0)

	for _, l := range labels {
		if l.Name == "__name__" {
			name = l.Value
		}
		newLabels = append(newLabels, Label{
			Key:   l.Name,
			Value: l.Value,
		})
	}

	m := Metric{
		name:   name,
		value:  value,
		ts:     time.UnixMilli(ts),
		labels: newLabels,
	}
	return m
}

func CopyMetricFromPrometheus(in *dto.MetricFamily) Metric {

	lbls := make([]Label, 0)
	foundName := false
	for _, v := range in.Metric[0].Label {
		if *v.Name == "__name__" {
			foundName = true
		}
		lbls = append(lbls, Label{
			Key:   *v.Name,
			Value: *v.Value,
		})
	}
	var val float64
	if in.Metric[0].Counter != nil {
		val = *in.Metric[0].Counter.Value
	} else if in.Metric[0].Gauge != nil {
		val = *in.Metric[0].Gauge.Value
	}
	m := Metric{
		name:   in.GetName(),
		value:  val,
		ts:     time.Now(),
		labels: lbls,
	}
	if !foundName {
		m.labels = append(m.labels, Label{
			Key:   "__name__",
			Value: m.name,
		})
	}

	return m
}

func CopyMetric(in Metric) Metric {
	return Metric{
		name:   in.Name(),
		value:  in.Value(),
		ts:     in.Timestamp(),
		labels: in.Labels(),
	}
}

func (m *Metric) Name() string {
	return m.name
}

func (m *Metric) Value() float64 {
	return m.value
}

func (m *Metric) Timestamp() time.Time {
	return m.ts
}

func (m *Metric) Labels() []Label {
	newLabels := make([]Label, len(m.labels))
	copy(newLabels, m.labels)
	return newLabels
}

func (m *Metric) FindLabel(name string) (Label, bool) {
	for _, l := range m.labels {
		if l.Key == name {
			return l, true
		}
	}
	return Label{}, false
}

func copyOrderedMap(in *orderedmap.OrderedMap) *orderedmap.OrderedMap {
	newMap := orderedmap.New()
	for _, k := range in.Keys() {
		v, _ := in.Get(k)
		newMap.Set(k, v)
	}
	return newMap
}
func copyMap(in map[string]string) map[string]string {
	newMap := make(map[string]string, 0)
	for k, v := range in {
		newMap[k] = v
	}
	return newMap
}

func copyLabels(in labels.Labels) *orderedmap.OrderedMap {
	newMap := orderedmap.New()
	for _, v := range in {
		newMap.Set(v.Name, v.Value)
	}
	return newMap
}

func CopyLabelSet(in model.LabelSet) *orderedmap.OrderedMap {
	newMap := orderedmap.New()
	for k, v := range in {
		newMap.Set(string(k), string(v))
	}
	return newMap
}
