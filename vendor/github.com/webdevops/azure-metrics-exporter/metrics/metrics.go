package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricHelpDefault = "Azure monitor insight metric"
)

type (
	MetricList struct {
		List map[string][]MetricRow
		Help map[string]string
	}

	MetricRow struct {
		Labels prometheus.Labels
		Value  float64
	}
)

func NewMetricList() *MetricList {
	list := MetricList{}
	list.List = map[string][]MetricRow{}
	list.Help = map[string]string{}
	return &list
}

func (l *MetricList) Add(name string, metric ...MetricRow) {
	if _, ok := l.List[name]; !ok {
		l.List[name] = []MetricRow{}
	}

	l.List[name] = append(l.List[name], metric...)
}

func (l *MetricList) GetMetricNames() (list []string) {
	for name := range l.List {
		list = append(list, name)
	}
	return
}

func (l *MetricList) SetMetricHelp(name, help string) {
	l.Help[name] = help
}

func (l *MetricList) GetMetricHelp(name string) string {
	if val, ok := l.Help[name]; ok {
		return val
	}
	return MetricHelpDefault
}

func (l *MetricList) GetMetricList(name string) []MetricRow {
	return l.List[name]
}

func (l *MetricList) GetMetricLabelNames(name string) []string {
	var list []string
	uniqueLabelMap := map[string]string{}

	for _, row := range l.List[name] {
		for labelName := range row.Labels {
			uniqueLabelMap[labelName] = labelName
		}
	}

	for labelName := range uniqueLabelMap {
		list = append(list, labelName)
	}

	return list
}
