package controller

import "github.com/prometheus/client_golang/prometheus"

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

type controllerCollector struct {
	ci                     ComponentInfos
	runningComponentsTotal *prometheus.Desc
}

type ComponentInfos interface {
	Components() []*ComponentNode
}

func NewControllerCollector(ci ComponentInfos) prometheus.Collector {
	return &controllerCollector{
		ci: ci,
		runningComponentsTotal: prometheus.NewDesc(
			"agent_component_controller_running_components_total",
			"Total number of running components.",
			[]string{"health_type"},
			nil,
		),
	}
}

func (cc *controllerCollector) Collect(ch chan<- prometheus.Metric) {
	componentsByHealth := make(map[string]int)

	for _, component := range cc.ci.Components() {
		health := component.CurrentHealth().Health.String()
		componentsByHealth[health]++
		component.register.Collect(ch)
	}

	for health, count := range componentsByHealth {
		ch <- prometheus.MustNewConstMetric(cc.runningComponentsTotal, prometheus.GaugeValue, float64(count), health)
	}
}

func (cc *controllerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.runningComponentsTotal
}
