package metrics

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/grafana/agent/pkg/agentflow/config"
	"github.com/grafana/agent/pkg/agentflow/types/actorstate"
	"github.com/grafana/agent/pkg/agentflow/types/exchange"
	"regexp"
)

type MetricFilter struct {
	cfg  config.MetricFilter
	self *actor.PID
	outs []*actor.PID
	name string
}

func (m *MetricFilter) AllowableInputs() []actorstate.InOutType {
	return []actorstate.InOutType{actorstate.Metrics}
}

func (m *MetricFilter) Output() actorstate.InOutType {
	return actorstate.Metrics
}

func NewMetricFilter(name string, cfg config.MetricFilter) (actorstate.FlowActor, error) {
	return &MetricFilter{
		cfg:  cfg,
		name: name,
	}, nil
}

func (m *MetricFilter) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case actorstate.Init:
		m.outs = msg.Children
	case actorstate.Start:
		m.self = c.Self()
	case []exchange.Metric:
		metrics := make([]exchange.Metric, 0)
		for _, metric := range msg {
			newM := m.match(metric)
			if newM == nil {
				continue
			}
			metrics = append(metrics, *newM)
		}
		for _, out := range m.outs {
			c.Send(out, metrics)
		}
	}
}

func (m *MetricFilter) Name() string {
	return m.name
}

func (m *MetricFilter) PID() *actor.PID {
	return m.self
}

func (m *MetricFilter) match(metric exchange.Metric) *exchange.Metric {
	if len(m.cfg.Filters) == 0 {
		return &metric
	}
	for _, f := range m.cfg.Filters {
		switch f.Action {
		case "drop_metric":
			matchedValue, found := metric.FindLabel(f.MatchField)
			if !found {
				continue
			}
			match, _ := regexp.MatchString(f.Regex, matchedValue.Value)
			if match {
				return nil
			}
			return &metric
		case "add_label":
			newMap := metric.Labels()
			newMap = append(newMap, exchange.Label{
				Key:   f.AddLabel,
				Value: f.AddValue,
			})
			newM := exchange.NewMetric(metric.Name(), metric.Value(), metric.Timestamp(), newMap)
			return &newM
		}
	}
	return &metric
}
