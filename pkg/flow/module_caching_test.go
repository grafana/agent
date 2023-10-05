package flow_test

// This file contains tests which verify that the Flow controller correctly updates and caches modules' arguments
// and exports in presence of multiple components.

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/service"
	cluster_service "github.com/grafana/agent/service/cluster"
	http_service "github.com/grafana/agent/service/http"
	otel_service "github.com/grafana/agent/service/otel"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	_ "github.com/grafana/agent/component/module/string"
)

func TestUpdates_EmptyModule(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)

	// There's an empty module in the config below, but the pipeline we test for propagation is not affected by it.
	config := `
	module.string "test" {
		content = ""
	}

	testcomponents.count "inc" {
		frequency = "10ms"
		max = 10
	}

	testcomponents.passthrough "inc_dep_1" {
		input = testcomponents.count.inc.count
		lag = "1ms"
	}

	testcomponents.passthrough "inc_dep_2" {
		input = testcomponents.passthrough.inc_dep_1.output
		lag = "1ms"
	}

	testcomponents.summation "sum" {
		input = testcomponents.passthrough.inc_dep_2.output
	}
`

	ctrl := flow.New(testOptions(t))
	f, err := flow.ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		ctrl.Run(ctx)
		close(done)
	}()
	defer func() {
		cancel()
		<-done
	}()

	require.Eventually(t, func() bool {
		export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
		return export.LastAdded == 10
	}, 3*time.Second, 10*time.Millisecond)
}

func TestUpdates_ThroughModule(t *testing.T) {
	// We use this module in a Flow config below.
	module := `
	argument "input" {
		optional = false
	}

	testcomponents.passthrough "pt" {
		input = argument.input.value
		lag = "1ms"
	}

	export "output" {
		value = testcomponents.passthrough.pt.output
	}
`

	// We send the count increments via module and to the summation component and verify that the updates propagate.
	config := `
	testcomponents.count "inc" {
		frequency = "10ms"
		max = 10
	}

	module.string "test" {
		content = ` + strconv.Quote(module) + `
		arguments {
			input = testcomponents.count.inc.count
		}
	}

	testcomponents.summation "sum" {
		input = module.string.test.exports.output
	}
`

	ctrl := flow.New(testOptions(t))
	f, err := flow.ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		ctrl.Run(ctx)
		close(done)
	}()
	defer func() {
		cancel()
		<-done
	}()

	require.Eventually(t, func() bool {
		export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
		return export.LastAdded == 10
	}, 3*time.Second, 10*time.Millisecond)
}

func testOptions(t *testing.T) flow.Options {
	t.Helper()
	s, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)

	clusterService, err := cluster_service.New(cluster_service.Options{
		Log:              s,
		EnableClustering: false,
		NodeName:         "test-node",
		AdvertiseAddress: "127.0.0.1:80",
	})
	require.NoError(t, err)

	otelService := otel_service.New(s)
	require.NotNil(t, otelService)

	return flow.Options{
		Logger:   s,
		DataPath: t.TempDir(),
		Reg:      nil,
		Services: []service.Service{
			http_service.New(http_service.Options{}),
			clusterService,
			otelService,
		},
	}
}

func getExport[T any](t *testing.T, ctrl *flow.Flow, moduleId string, nodeId string) T {
	t.Helper()
	info, err := ctrl.GetComponent(component.ID{
		ModuleID: moduleId,
		LocalID:  nodeId,
	}, component.InfoOptions{
		GetHealth:    true,
		GetArguments: true,
		GetExports:   true,
		GetDebugInfo: true,
	})
	require.NoError(t, err)
	return info.Exports.(T)
}

func verifyNoGoroutineLeaks(t *testing.T) {
	t.Helper()
	goleak.VerifyNone(
		t,
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreTopFunction("go.opentelemetry.io/otel/sdk/trace.(*batchSpanProcessor).processQueue"),
	)
}
