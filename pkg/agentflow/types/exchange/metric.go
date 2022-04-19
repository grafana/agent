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
	labels   *orderedmap.OrderedMap
	metadata *orderedmap.OrderedMap
}

func NewMetric(name string, value float64, ts time.Time, labels *orderedmap.OrderedMap, metadata *orderedmap.OrderedMap) Metric {
	m := Metric{
		name:     name,
		value:    value,
		ts:       ts,
		labels:   labels,
		metadata: metadata,
	}
	if m.labels == nil {
		m.labels = orderedmap.New()
	}
	if m.metadata == nil {
		m.metadata = orderedmap.New()
	}
	m.labels.Set("__name__", name)
	return m
}

func NewMetricFromPromMetric(ts int64, value float64, labels labels.Labels) Metric {
	name := ""
	newLabels := orderedmap.New()

	for _, l := range labels {
		if l.Name == "__name__" {
			name = l.Value
		}
		newLabels.Set(l.Name, l.Value)
	}

	m := Metric{
		name:   name,
		value:  value,
		ts:     time.UnixMilli(ts),
		labels: newLabels,
	}
	if m.labels == nil {
		m.labels = orderedmap.New()
	}
	if m.metadata == nil {
		m.metadata = orderedmap.New()
	}
	return m
}

func CopyMetricFromPrometheus(in *dto.MetricFamily) Metric {

	lbls := orderedmap.New()
	for _, v := range in.Metric[0].Label {
		lbls.Set(*v.Name, *v.Value)
	}
	var val float64
	if in.Metric[0].Counter != nil {
		val = *in.Metric[0].Counter.Value
	} else if in.Metric[0].Gauge != nil {
		val = *in.Metric[0].Gauge.Value
	}
	m := Metric{
		name:     in.GetName(),
		value:    val,
		ts:       time.Now(),
		labels:   lbls,
		metadata: orderedmap.New(),
	}
	m.labels.Set("__name__", in.GetName())
	return m
}

func CopyMetric(in Metric) Metric {
	return Metric{
		name:     in.Name(),
		value:    in.Value(),
		ts:       in.Timestamp(),
		labels:   in.Labels(),
		metadata: in.Metadata(),
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

func (m *Metric) Labels() *orderedmap.OrderedMap {
	return copyMap(m.labels)
}

func (m *Metric) Metadata() *orderedmap.OrderedMap {
	return copyMap(m.metadata)
}

func copyMap(in *orderedmap.OrderedMap) *orderedmap.OrderedMap {
	newMap := orderedmap.New()
	for _, k := range in.Keys() {
		v, _ := in.Get(k)
		newMap.Set(k, v)
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
