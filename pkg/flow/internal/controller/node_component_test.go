package controller

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalID(t *testing.T) {
	mo := getManagedOptions(ComponentGlobals{
		DataPath:     "/data/",
		ControllerID: "module.file",
		NewModuleController: func(id string, availableServices []string) ModuleController {
			return nil
		},
	}, &ComponentNode{
		nodeID:   "local.id",
		globalID: "module.file/local.id",
	})
	require.Equal(t, "/data/module.file/local.id", filepath.ToSlash(mo.DataPath))
}

func TestLocalID(t *testing.T) {
	mo := getManagedOptions(ComponentGlobals{
		DataPath:     "/data/",
		ControllerID: "",
		NewModuleController: func(id string, availableServices []string) ModuleController {
			return nil
		},
	}, &ComponentNode{
		nodeID:   "local.id",
		globalID: "local.id",
	})
	require.Equal(t, "/data/local.id", filepath.ToSlash(mo.DataPath))
}
