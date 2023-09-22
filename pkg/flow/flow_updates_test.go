package flow

import (
	"context"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testUpdatesFile = `
	testcomponents.count "inc" {
		frequency = "10ms"
		max = 10
	}

	testcomponents.passthrough "inc_dep_1" {
		input = testcomponents.count.inc.count
		//lag = "10ms"
	}

	testcomponents.passthrough "inc_dep_2" {
		input = testcomponents.passthrough.inc_dep_1.output
		//lag = "10ms"
	}

	testcomponents.summation "sum" {
		input = testcomponents.passthrough.inc_dep_2.output
	}
`

func TestController_Updates(t *testing.T) {
	ctrl := New(testOptions(t))

	// Use testUpdatesFile from graph_builder_test.go.
	f, err := ParseSource(t.Name(), []byte(testUpdatesFile))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)
	require.Len(t, ctrl.loader.Components(), 4)

	ctx, cancel := context.WithCancel(context.Background())
	go ctrl.Run(ctx)

	// Wait for the updates to propagate
	require.Eventually(t, func() bool {
		_, out := getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
		return out.(testcomponents.SummationExports).Output == 55
	}, 5*time.Second, 10*time.Millisecond)

	in, out := getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_1")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.passthrough.inc_dep_2")
	require.Equal(t, "10", in.(testcomponents.PassthroughConfig).Input)
	require.Equal(t, "10", out.(testcomponents.PassthroughExports).Output)

	in, out = getFields(t, ctrl.loader.Graph(), "testcomponents.summation.sum")
	require.Equal(t, 10, in.(testcomponents.SummationConfig).Input)
	require.Equal(t, 55, out.(testcomponents.SummationExports).Output)

	cancel()
}

//TODO(thampiotr): add test with modules
