package pogo

import (
	dto "github.com/prometheus/client_model/go"
	"time"
)

type Metric struct {
	name     string
	value    float64
	ts       time.Time
	labels   map[string]string
	metadata map[string]string
}

func NewMetric(name string, value float64, ts time.Time, labels map[string]string, metadata map[string]string) Metric {
	m := Metric{
		name:     name,
		value:    value,
		ts:       ts,
		labels:   labels,
		metadata: metadata,
	}
	if m.labels == nil {
		m.labels = make(map[string]string)
	}
	if m.metadata == nil {
		m.labels = make(map[string]string)
	}
	m.labels["__name__"] = name
	return m
}

func CopyMetricFromPrometheus(in *dto.MetricFamily) Metric {

	lbls := make(map[string]string)
	for _, v := range in.Metric[0].Label {
		lbls[*v.Name] = *v.Value
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
		metadata: nil,
	}
	m.labels["__name__"] = in.GetName()
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

func (m *Metric) Labels() map[string]string {
	return copyMap(m.labels)
}

func (m *Metric) Metadata() map[string]string {
	return copyMap(m.metadata)
}

func copyMap(in map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range in {
		newMap[k] = v
	}
	return newMap
}
