package flow

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/flow/internal/worker"
	"github.com/stretchr/testify/require"
)

func TestController_Updates(t *testing.T) {
	// Simple pipeline with a minimal lag
	config := `
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

	ctrl := newTestController(t)

	// Use testUpdatesFile from graph_builder_test.go.
	f, err := ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go ctrl.Run(ctx)

	// Wait for the updates to propagate
	require.Eventually(t, func() bool {
		_, out := getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
		return out.(testcomponents.SummationExports).LastAdded == 10
	}, 3*time.Second, 10*time.Millisecond)

	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_1")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_2")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
	require.Equal(t, 10, in.(testcomponents.SummationConfig).Input)

	// Since the lag is minimal, all updates will arrive to the final node.
	require.Equal(t, 55, out.(testcomponents.SummationExports).Sum)

	cancel()
}

func TestController_Updates_WithLag(t *testing.T) {
	// Simple pipeline with some lag
	config := `
	testcomponents.count "inc" {
		frequency = "10ms"
		max = 10
	}

	testcomponents.passthrough "inc_dep_1" {
		input = testcomponents.count.inc.count
		lag = "10ms"
	}

	testcomponents.passthrough "inc_dep_2" {
		input = testcomponents.passthrough.inc_dep_1.output
		lag = "10ms"
	}

	testcomponents.summation "sum" {
		input = testcomponents.passthrough.inc_dep_2.output
	}
`

	ctrl := newTestController(t)

	// Use testUpdatesFile from graph_builder_test.go.
	f, err := ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go ctrl.Run(ctx)

	// Wait for the updates to propagate
	require.Eventually(t, func() bool {
		_, out := getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
		return out.(testcomponents.SummationExports).LastAdded == 10
	}, 3*time.Second, 10*time.Millisecond)

	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_1")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_2")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, _ = getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
	require.Equal(t, 10, in.(testcomponents.SummationConfig).Input)

	cancel()
}

func TestController_Updates_WithOtherLaggingPipeline(t *testing.T) {
	// Another pipeline exists with a significant lag.
	config := `
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

	testcomponents.count "inc_2" {
		frequency = "10ms"
		max = 10
	}

	testcomponents.passthrough "inc_dep_slow" {
		input = testcomponents.count.inc_2.count
		lag = "1s"
	}
`

	ctrl := newTestController(t)

	// Use testUpdatesFile from graph_builder_test.go.
	f, err := ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go ctrl.Run(ctx)

	// Wait for the updates to propagate
	require.Eventually(t, func() bool {
		_, out := getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
		return out.(testcomponents.SummationExports).LastAdded == 10
	}, 3*time.Second, 10*time.Millisecond)

	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_1")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_2")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
	require.Equal(t, 10, in.(testcomponents.SummationConfig).Input)

	// Since the actual lag should be minimal, all updates should arrive to the final node.
	require.Equal(t, 55, out.(testcomponents.SummationExports).Sum)

	cancel()
}

func TestController_Updates_WithLaggingComponent(t *testing.T) {
	// Part of the pipeline has a significant lag.
	config := `
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

	testcomponents.passthrough "inc_dep_slow" {
		input = testcomponents.count.inc.count
		lag = "1s"
	}
`

	ctrl := newTestController(t)

	// Use testUpdatesFile from graph_builder_test.go.
	f, err := ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go ctrl.Run(ctx)

	// Wait for the updates to propagate
	require.Eventually(t, func() bool {
		_, out := getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
		return out.(testcomponents.SummationExports).LastAdded == 10
	}, 3*time.Second, 10*time.Millisecond)

	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_1")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_2")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
	require.Equal(t, 10, in.(testcomponents.SummationConfig).Input)

	// Since the actual lag should be minimal, all updates should arrive to the final node.
	require.Equal(t, 55, out.(testcomponents.SummationExports).Sum)

	cancel()
}

func newTestController(t *testing.T) *Flow {
	return newController(controllerOptions{
		Options:        testOptions(t),
		ModuleRegistry: newModuleRegistry(),
		IsModule:       false,
		// Make sure that we have consistent number of workers for tests to make them deterministic.
		WorkerPool: worker.NewShardedWorkerPool(4, 100),
	})
}
