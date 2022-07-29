package controller

import "github.com/prometheus/client_golang/prometheus"

// ControllerMetrics contains the metrics for components controller
type ControllerMetrics struct {
	r prometheus.Registerer

	controllerEvaluation    prometheus.Gauge
	componentEvaluationTime prometheus.Histogram
	runningComponentTotal   *prometheus.CounterVec
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

	cm.runningComponentTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "agent_component_running_component_total",
		Help: "Total of running components",
	}, []string{"health_type"})

	if r != nil {
		r.MustRegister(
			cm.controllerEvaluation,
			cm.componentEvaluationTime,
			cm.runningComponentTotal,
		)
	}
	return &cm
}
