package exporters

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/agent/pkg/integrations/v2/app_o11y_receiver/models"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/assert"
)

type metricAssertion struct {
	name  string
	value float64
}

func testcase(t *testing.T, payload models.Payload, assertions []metricAssertion) {
	ctx := context.Background()

	reg := prometheus.NewRegistry()

	exporter := NewReceiverMetricsExporter(ReceiverMetricsExporterConfig{Reg: reg})

	err := exporter.Export(ctx, payload)
	assert.NoError(t, err)

	metrics, err := reg.Gather()
	assert.NoError(t, err)

	for _, assertion := range assertions {
		found := false
		for _, metric := range metrics {
			if *metric.Name == assertion.name {
				found = true
				assert.Len(t, metric.Metric, 1)
				val := metric.Metric[0].Counter.Value
				assert.Equal(t, assertion.value, *val)
				break
			}
		}
		if !found {
			assert.Fail(t, fmt.Sprintf("metric [%s] not found", assertion.name))
		}
	}
}

func TestReceiverMetricsExport(t *testing.T) {
	var payload models.Payload
        payload.Logs = make([]models.Log, 2)
        payload.Measurements = make([]models.Measurement, 3)
        payload.Exceptions = make([]models.Exception, 4)
	testcase(t, payload, []metricAssertion{
		{
			name:  "app_o11y_receiver_total_logs",
			value: 2,
		},
		{
			name:  "app_o11y_receiver_total_measurements",
			value: 3,
		},
		{
			name:  "app_o11y_receiver_total_exceptions",
			value: 4,
		},
	})
}

func TestReceiverMetricsExportLogsOnly(t *testing.T) {
	var payload models.Payload
	payload.Logs = []models.Log{
		{},
		{},
	}
	testcase(t, payload, []metricAssertion{
		{
			name:  "app_o11y_receiver_total_logs",
			value: 2,
		},
		{
			name:  "app_o11y_receiver_total_measurements",
			value: 0,
		},
		{
			name:  "app_o11y_receiver_total_exceptions",
			value: 0,
		},
	})
}

func TestReceiverMetricsExportExceptionsOnly(t *testing.T) {
	var payload models.Payload
	payload.Exceptions = []models.Exception{
		{},
		{},
		{},
		{},
	}
	testcase(t, payload, []metricAssertion{
		{
			name:  "app_o11y_receiver_total_logs",
			value: 0,
		},
		{
			name:  "app_o11y_receiver_total_measurements",
			value: 0,
		},
		{
			name:  "app_o11y_receiver_total_exceptions",
			value: 4,
		},
	})
}

func TestReceiverMetricsExportMeasurementsOnly(t *testing.T) {
	var payload models.Payload
	payload.Measurements = []models.Measurement{
		{},
		{},
		{},
	}
	testcase(t, payload, []metricAssertion{
		{
			name:  "app_o11y_receiver_total_logs",
			value: 0,
		},
		{
			name:  "app_o11y_receiver_total_measurements",
			value: 3,
		},
		{
			name:  "app_o11y_receiver_total_exceptions",
			value: 0,
		},
	})
}
