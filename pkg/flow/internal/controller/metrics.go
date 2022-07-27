package controller

import "github.com/prometheus/client_golang/prometheus"

var (
	controllerEvaluation     prometheus.Gauge
	componentEvaluationTime  prometheus.Histogram
	runningComponentStatus   prometheus.GaugeVec
	evaluatedComponentStatus prometheus.GaugeVec
)

func initControllerMetrics(r prometheus.Registerer) {
	controllerEvaluation := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "agent_component_controller_evaluating",
		Help: "Tracks if the controller is currently in the middle of a graph evaluation",
	})
	componentEvaluationTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "agent_component_evaluation_seconds",
			Help: "Time spent performing component evaluation",
		},
	)

	componentLabels := []string{"id", "health_type"}
	runningComponentStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "agent_component_running_component_status",
		Help: "Status of a running component",
	}, componentLabels)

	evaluatedComponentStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "agent_component_evaluated_component_status",
		Help: "Status of an evaluated component",
	}, componentLabels)

	if r != nil {
		r.MustRegister(
			controllerEvaluation,
			componentEvaluationTime,
			runningComponentStatus,
			evaluatedComponentStatus,
		)
	}

}
