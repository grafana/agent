package exporters

import (
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/utils"
	"github.com/prometheus/client_golang/prometheus"
)

// ReceiverMetricsExporterConfig contains options for the ReceiverMetricsExporter
type ReceiverMetricsExporterConfig struct {
	Reg *prometheus.Registry
}

// ReceiverMetricsExporter is a app o11y receiver exporter that will capture metrics
// about counts of logs, exceptions, measurements, traces being ingested
type ReceiverMetricsExporter struct {
	totalLogs         prometheus.Counter
	totalMeasurements prometheus.Counter
	totalExceptions   prometheus.Counter
}

// NewReceiverMetricsExporter creates a new ReceiverMetricsExporter
func NewReceiverMetricsExporter(conf ReceiverMetricsExporterConfig) AppO11yReceiverExporter {
	exp := &ReceiverMetricsExporter{
		totalLogs: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: utils.MetricsNamespace,
			Name:      "total_logs",
			Help:      "Total number of ingested logs",
		}),
		totalMeasurements: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: utils.MetricsNamespace,
			Name:      "total_measurements",
			Help:      "Total number of ingested measurements",
		}),
		totalExceptions: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: utils.MetricsNamespace,
			Name:      "total_exceptions",
			Help:      "Total number of ingested exceptions",
		}),
	}

	conf.Reg.MustRegister(exp.totalLogs, exp.totalExceptions, exp.totalMeasurements)

	return exp
}

// Name of the exporter, for logging purposes
func (re *ReceiverMetricsExporter) Name() string {
	return "receiver metrics exporter"
}

// Export implements the AppDataExporter interface
func (re *ReceiverMetricsExporter) Export(payload models.Payload) error {
	re.totalExceptions.Add(float64(len(payload.Exceptions)))
	re.totalLogs.Add(float64(len(payload.Logs)))
	re.totalMeasurements.Add(float64(len(payload.Measurements)))
	return nil
}
