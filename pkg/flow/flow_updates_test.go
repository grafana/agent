package flow

import (
	"context"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"testing"
	"time"

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

	ctrl := New(testOptions(t))

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

	ctrl := New(testOptions(t))

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

	// Since the lag is significant, we can be sure that some updates will be missed:
	require.NotEqual(t, 55, out.(testcomponents.SummationExports).Sum)

	cancel()
}

func TestController_Updates_WithOtherLaggingPipeline(t *testing.T) {
	t.Skipf("This test will frequently fail right now, as each partilal graph evaluation including the lagging" +
		" components, will block evaluation of other components.")

	//TODO(thampiotr): Parallelise graph evaluation to avoid this issue.

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

	ctrl := New(testOptions(t))

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
	}, 2*time.Second, 10*time.Millisecond)

	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_1")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_2")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
	require.Equal(t, 10, in.(testcomponents.SummationConfig).Input)

	cancel()
}

//TODO(thampiotr): add correctness test with modules
