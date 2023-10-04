package controller

import (
	"github.com/prometheus/client_golang/prometheus"
)

// controllerMetrics contains the metrics for components controller
type controllerMetrics struct {
	controllerEvaluation    prometheus.Gauge
	componentEvaluationTime prometheus.Histogram
	dependenciesWaitTime    prometheus.Histogram
	evaluationQueueSize     prometheus.Gauge
}

// newControllerMetrics inits the metrics for the components controller
func newControllerMetrics(id string) *controllerMetrics {
	cm := &controllerMetrics{}

	cm.controllerEvaluation = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "agent_component_controller_evaluating",
		Help:        "Tracks if the controller is currently in the middle of a graph evaluation",
		ConstLabels: map[string]string{"controller_id": id},
	})

	cm.componentEvaluationTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:        "agent_component_evaluation_seconds",
			Help:        "Time spent performing component evaluation",
			ConstLabels: map[string]string{"controller_id": id},
		},
	)
	cm.dependenciesWaitTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:        "agent_component_dependencies_wait_seconds",
			Help:        "Time spent by components waiting to be evaluated after their dependency is updated.",
			ConstLabels: map[string]string{"controller_id": id},
		},
	)

	cm.evaluationQueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "agent_component_evaluation_queue_size",
		Help:        "Tracks the number of components waiting to be evaluated in the worker pool",
		ConstLabels: map[string]string{"controller_id": id},
	})

	return cm
}

func (cm *controllerMetrics) Collect(ch chan<- prometheus.Metric) {
	cm.componentEvaluationTime.Collect(ch)
	cm.controllerEvaluation.Collect(ch)
	cm.dependenciesWaitTime.Collect(ch)
	cm.evaluationQueueSize.Collect(ch)
}

func (cm *controllerMetrics) Describe(ch chan<- *prometheus.Desc) {
	cm.componentEvaluationTime.Describe(ch)
	cm.controllerEvaluation.Describe(ch)
	cm.dependenciesWaitTime.Describe(ch)
	cm.evaluationQueueSize.Describe(ch)
}

type controllerCollector struct {
	l                      *Loader
	runningComponentsTotal *prometheus.Desc
}

func newControllerCollector(l *Loader, id string) *controllerCollector {
	return &controllerCollector{
		l: l,
		runningComponentsTotal: prometheus.NewDesc(
			"agent_component_controller_running_components",
			"Total number of running components.",
			[]string{"health_type"},
			map[string]string{"controller_id": id},
		),
	}
}

func (cc *controllerCollector) Collect(ch chan<- prometheus.Metric) {
	componentsByHealth := make(map[string]int)

	for _, component := range cc.l.Components() {
		health := component.CurrentHealth().Health.String()
		componentsByHealth[health]++
		component.registry.Collect(ch)
	}

	for health, count := range componentsByHealth {
		ch <- prometheus.MustNewConstMetric(cc.runningComponentsTotal, prometheus.GaugeValue, float64(count), health)
	}
}

func (cc *controllerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.runningComponentsTotal
}
