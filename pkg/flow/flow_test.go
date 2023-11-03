package flow

import (
	"context"
	"os"
	"testing"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/flow/internal/controller"
	"github.com/grafana/agent/pkg/flow/internal/dag"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var testFile = `
	testcomponents.tick "ticker" {
		frequency = "1s"
	}

	testcomponents.passthrough "static" {
		input = "hello, world!"
	}

	testcomponents.passthrough "ticker" {
		input = testcomponents.tick.ticker.tick_time
	}

	testcomponents.passthrough "forwarded" {
		input = testcomponents.passthrough.ticker.output
	}
`

func TestController_LoadSource_Evaluation(t *testing.T) {
	defer verifyNoGoroutineLeaks(t)
	ctrl := New(testOptions(t))
	defer cleanUpController(ctrl)

	// Use testFile from graph_builder_test.go.
	f, err := ParseSource(t.Name(), []byte(testFile))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)
	require.Len(t, ctrl.loader.Components(), 4)

	// Check the inputs and outputs of things that should be immediately resolved
	// without having to run the components.
	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.static")
	require.Equal(t, "hello, world!", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "hello, world!", out.(testcomponents.PassthroughExports).Output)
}

func getFields(t *testing.T, g *dag.Graph, nodeID string) (component.Arguments, component.Exports) {
	t.Helper()

	n := g.GetByID(nodeID)
	require.NotNil(t, n, "couldn't find node %q in graph", nodeID)

	uc := n.(*controller.ComponentNode)
	return uc.Arguments(), uc.Exports()
}

func testOptions(t *testing.T) Options {
	t.Helper()

	s, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)

	return Options{
		Logger:   s,
		DataPath: t.TempDir(),
		Reg:      nil,
	}
}

func cleanUpController(ctrl *Flow) {
	// To avoid leaking goroutines and clean-up, we need to run and shut down the controller.
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		ctrl.Run(ctx)
		close(done)
	}()
	cancel()
	<-done
}

func verifyNoGoroutineLeaks(t *testing.T) {
	t.Helper()
	goleak.VerifyNone(
		t,
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreTopFunction("go.opentelemetry.io/otel/sdk/trace.(*batchSpanProcessor).processQueue"),
	)
}
