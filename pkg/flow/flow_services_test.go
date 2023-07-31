package flow

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow/internal/testservices"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/agent/service"
	"github.com/stretchr/testify/require"
)

func TestServices(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		startedSvc = util.NewWaitTrigger()

		svc = &testservices.Fake{
			RunFunc: func(ctx context.Context, _ service.Host) error {
				startedSvc.Trigger()

				<-ctx.Done()
				return nil
			},
		}
	)

	opts := testOptions(t)
	opts.Services = append(opts.Services, svc)

	ctrl := New(opts)
	require.NoError(t, ctrl.LoadFile(makeEmptyFile(t), nil))

	// Start the controller. This should cause our service to run.
	go ctrl.Run(ctx)

	require.NoError(t, startedSvc.Wait(5*time.Second), "Service did not start")
}

func TestFlow_GetServiceConsumers(t *testing.T) {
	var (
		svcA = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{
					Name: "svc_a",
				}
			},
		}

		svcB = &testservices.Fake{
			DefinitionFunc: func() service.Definition {
				return service.Definition{
					Name:      "svc_b",
					DependsOn: []string{"svc_a"},
				}
			},
		}
	)

	opts := testOptions(t)
	opts.Services = append(opts.Services, svcA, svcB)

	ctrl := New(opts)
	require.NoError(t, ctrl.LoadFile(makeEmptyFile(t), nil))

	consumers := ctrl.GetServiceConsumers("svc_a")
	require.Equal(t, []any{svcB}, consumers)
}

func makeEmptyFile(t *testing.T) *File {
	t.Helper()

	f, err := ReadFile(t.Name(), nil)
	require.NoError(t, err)
	require.NotNil(t, f)

	return f
}
