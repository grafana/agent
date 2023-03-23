package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/auth"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func TestAuth(t *testing.T) {
	var (
		waitCreated = util.NewWaitTrigger()
		onCreated   = func() {
			waitCreated.Trigger()
		}
	)

	// Create and start our Flow component. We then wait for it to export a
	// consumer that we can send data to.
	te := newTestEnvironment(t, onCreated)
	te.Start(fakeAuthArgs{})

	require.NoError(t, waitCreated.Wait(time.Second), "extension never created")
}

type testEnvironment struct {
	t *testing.T

	Controller *componenttest.Controller
}

func newTestEnvironment(t *testing.T, onCreated func()) *testEnvironment {
	t.Helper()

	reg := component.Registration{
		Name:    "testcomponent",
		Args:    fakeAuthArgs{},
		Exports: otelcol.ConsumerExports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			factory := otelcomponent.NewExtensionFactory(
				"testcomponent",
				func() otelconfig.Extension { return nil },
				func(
					_ context.Context,
					_ otelcomponent.ExtensionCreateSettings,
					_ otelconfig.Extension,
				) (otelcomponent.Extension, error) {

					onCreated()
					return nil, nil
				}, otelcomponent.StabilityLevelUndefined,
			)

			return auth.New(opts, factory, args.(auth.Arguments))
		},
	}

	return &testEnvironment{
		t:          t,
		Controller: componenttest.NewControllerFromReg(util.TestLogger(t), reg),
	}
}

func (te *testEnvironment) Start(args component.Arguments) {
	go func() {
		ctx := componenttest.TestContext(te.t)
		err := te.Controller.Run(ctx, args)
		require.NoError(te.t, err, "failed to run component")
	}()
}

type fakeAuthArgs struct {
}

var _ auth.Arguments = fakeAuthArgs{}

func (fa fakeAuthArgs) Convert() (otelconfig.Extension, error) {
	settings := otelconfig.NewExtensionSettings(otelconfig.NewComponentID("testcomponent"))
	return &settings, nil
}

func (fa fakeAuthArgs) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

func (fa fakeAuthArgs) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
