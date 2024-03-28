package flow_test

import (
	"context"
	"github.com/grafana/agent/internal/component"
	"github.com/grafana/agent/internal/flow"
	"github.com/grafana/agent/internal/flow/internal/controller"
	"github.com/grafana/agent/internal/flow/internal/dag"
	"testing"
	"time"

	_ "github.com/grafana/agent/internal/component/all"
	"github.com/grafana/agent/internal/flow/internal/worker"
	"github.com/stretchr/testify/require"
)

func TestPG(t *testing.T) {
	config := `
logging {
    level = "debug"
}

prometheus.exporter.postgres "pg" {
    data_source_names = [
        "postgresql://user1:password1@localhost:5432/demo1?sslmode=disable",
        "postgresql://user2:password2@localhost:5433/demo2?sslmode=disable",
    ]
}
`

	ctrl := newTestController(t)

	// Use testUpdatesFile from graph_builder_test.go.
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

	time.Sleep(90 * time.Minute)
}

func newTestController(t *testing.T) *flow.Flow {
	return flow.NewController(flow.ControllerOptions{
		Options:        testOptions(t),
		ModuleRegistry: flow.NewModuleRegistry(),
		IsModule:       false,
		// Make sure that we have consistent number of workers for tests to make them deterministic.
		WorkerPool: worker.NewFixedWorkerPool(4, 100),
	})
}

func getFields(t *testing.T, g *dag.Graph, nodeID string) (component.Arguments, component.Exports) {
	t.Helper()

	n := g.GetByID(nodeID)
	require.NotNil(t, n, "couldn't find node %q in graph", nodeID)

	uc := n.(*controller.BuiltinComponentNode)
	return uc.Arguments(), uc.Exports()
}
