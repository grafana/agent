package controller

import "github.com/prometheus/client_golang/prometheus"

// ControllerMetrics contains the metrics for components controller
type ControllerMetrics struct {
	r prometheus.Registerer

	controllerEvaluation     prometheus.Gauge
	componentEvaluationTime  prometheus.Histogram
	runningHealthyComponents prometheus.Gauge
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

	cm.runningHealthyComponents = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_component_running_healthy_components",
		Help: "Number of running healthy components",
	})

	if r != nil {
		r.MustRegister(
			cm.controllerEvaluation,
			cm.componentEvaluationTime,
			cm.runningHealthyComponents,
		)
	}
	return &cm
}
