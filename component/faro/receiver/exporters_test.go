package receiver

import (
	"context"
	"strings"
	"testing"

	"github.com/grafana/agent/component/faro/receiver/internal/payload"
	"github.com/prometheus/client_golang/prometheus"
	promtestutil "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

var metricNames = []string{
	"logs_total",
	"measurements_total",
	"exceptions_total",
	"events_total",
}

func Test_metricsExporter_Export(t *testing.T) {
	var (
		reg = prometheus.NewRegistry()
		exp = newMetricsExporter(reg)
	)

	expect := `
		# HELP logs_total Total number of ingested logs
		# TYPE logs_total counter
		logs_total 2

		# HELP measurements_total Total number of ingested measurements
		# TYPE measurements_total counter
		measurements_total 3

		# HELP exceptions_total Total number of ingested exceptions
		# TYPE exceptions_total counter
		exceptions_total 4

		# HELP events_total Total number of ingested events
		# TYPE events_total counter
		events_total 5
	`

	p := payload.Payload{
		Logs:         make([]payload.Log, 2),
		Measurements: make([]payload.Measurement, 3),
		Exceptions:   make([]payload.Exception, 4),
		Events:       make([]payload.Event, 5),
	}
	require.NoError(t, exp.Export(context.Background(), p))

	err := promtestutil.CollectAndCompare(reg, strings.NewReader(expect), metricNames...)
	require.NoError(t, err)
}
