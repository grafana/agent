package controller

import (
	"testing"

	"github.com/grafana/agent/component"
	"github.com/stretchr/testify/require"
)

func TestGlobalID(t *testing.T) {
	mo := getManagedOptions(ComponentGlobals{
		DataPath:       "/data/",
		HTTPPathPrefix: "/http/",
		ControllerID:   "module.file",
		NewModuleController: func(id string) component.ModuleController {
			return nil
		},
	}, &ComponentNode{
		nodeID: "local.id",
	})
	require.True(t, mo.HTTPPath == "/http/module.file/local.id/")
	require.True(t, mo.DataPath == "/data/module.file/local.id")
}

func TestLocalID(t *testing.T) {
	mo := getManagedOptions(ComponentGlobals{
		DataPath:       "/data/",
		HTTPPathPrefix: "/http/",
		ControllerID:   "",
		NewModuleController: func(id string) component.ModuleController {
			return nil
		},
	}, &ComponentNode{
		nodeID: "local.id",
	})
	require.True(t, mo.HTTPPath == "/http/local.id/")
	require.True(t, mo.DataPath == "/data/local.id")
}
