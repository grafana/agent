package extension_test

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/otelcol"
	"github.com/grafana/agent/component/otelcol/extension"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconfig "go.opentelemetry.io/collector/config"
)

func TestExtension(t *testing.T) {
	var (
		waitCreated = util.NewWaitTrigger()
		onCreated   = func() {
			waitCreated.Trigger()
		}
	)

	// Create and start our Flow component. We then wait for it to export a
	// consumer that we can send data to.
	te := newTestEnvironment(t, onCreated)
	te.Start(fakeExtensionArgs{})

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
		Args:    fakeExtensionArgs{},
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

			return extension.New(opts, factory, args.(extension.Arguments))
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

type fakeExtensionArgs struct {
}

var _ extension.Arguments = fakeExtensionArgs{}

func (fa fakeExtensionArgs) Convert() (otelconfig.Extension, error) {
	settings := otelconfig.NewExtensionSettings(otelconfig.NewComponentID("testcomponent"))
	return &settings, nil
}

func (fa fakeExtensionArgs) Extensions() map[otelconfig.ComponentID]otelcomponent.Extension {
	return nil
}

func (fa fakeExtensionArgs) Exporters() map[otelconfig.DataType]map[otelconfig.ComponentID]otelcomponent.Exporter {
	return nil
}
