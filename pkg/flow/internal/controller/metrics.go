package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

// ControllerMetrics contains the metrics for components controller
type ControllerMetrics struct {
	r prometheus.Registerer

	controllerEvaluation    prometheus.Gauge
	componentEvaluationTime prometheus.Histogram
}

// NewControllerMetrics inits the metrics for the components controller
func NewControllerMetrics(r prometheus.Registerer) *ControllerMetrics {
	cm := ControllerMetrics{r: r}

	cm.controllerEvaluation = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_component_controller_evaluating",
		Help: "Tracks if the controller is currently in the middle of a graph evaluation",
	})

	cm.componentEvaluationTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "agent_component_evaluation_seconds",
			Help: "Time spent performing component evaluation",
		},
	)

	if r != nil {
		r.MustRegister(
			cm.controllerEvaluation,
			cm.componentEvaluationTime,
		)
	}
	return &cm
}

type ControllerCollector struct {
	mut                    sync.Mutex
	l                      []*Loader
	runningComponentsTotal *prometheus.Desc
}

func NewControllerCollector() *ControllerCollector {
	return &ControllerCollector{
		l: make([]*Loader, 0),
		runningComponentsTotal: prometheus.NewDesc(
			"agent_component_controller_running_components_total",
			"Total number of running components.",
			[]string{"health_type"},
			nil,
		),
	}
}

func (cc *ControllerCollector) AddLoader(lod *Loader) {
	cc.mut.Lock()
	defer cc.mut.Unlock()
	cc.l = append(cc.l, lod)
}

func (cc *ControllerCollector) Collect(ch chan<- prometheus.Metric) {
	cc.mut.Lock()
	defer cc.mut.Unlock()
	componentsByHealth := make(map[string]int)
	for _, l := range cc.l {
		for _, component := range l.Components() {
			health := component.CurrentHealth().Health.String()
			componentsByHealth[health]++
			component.register.Collect(ch)
		}
	}

	for health, count := range componentsByHealth {
		ch <- prometheus.MustNewConstMetric(cc.runningComponentsTotal, prometheus.GaugeValue, float64(count), health)
	}
}

func (cc *ControllerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.runningComponentsTotal
}
