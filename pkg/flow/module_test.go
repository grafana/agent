package flow

import (
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestLoadConfig(t *testing.T) {
	controller := newDelegate("test", createOptions(t))
	err := controller.LoadConfig([]byte("bad config"), component.Options{}, nil, func(exports map[string]any) {})
	require.Error(t, err)
}

func createOptions(t *testing.T) *ModuleOptions {
	l := util.TestFlowLogger(t)
	cl, err := cluster.New(l, prometheus.DefaultRegisterer, false, "", "", "")
	require.NoError(t, err)
	return &ModuleOptions{
		Logger:    l,
		Tracer:    trace.NewNoopTracerProvider(),
		Clusterer: cl,
		Reg:       prometheus.NewRegistry(),
	}
}
