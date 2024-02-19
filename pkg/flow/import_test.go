package flow_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/internal/testcomponents"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/grafana/agent/service"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"

	_ "github.com/grafana/agent/component/module/string"
)

// updateFile is used to change the content of a file used by a module
type updateFile struct {
	name         string   // name of the file which should be updated (module | nested_module)
	updateConfig []string // new module config which should be used
}

func TestImport(t *testing.T) {
	modules := loadModules(t)
	testCases := []struct {
		name         string
		config       []string    // root config that the controller should load
		module       []string    // module (file) that the root config can import
		nestedModule []string    // module (file) that the module can import
		update       *updateFile // update can update the module or the nested_module file with new content
	}{
		{
			name:   "Import passthrough module.",
			config: []string{"root_import_a"},
			module: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:   "Import passthrough module in a declare.",
			config: []string{"root_import_a_in_declare_b"},
			module: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:   "Import passthrough module; instantiate imported declare in a declare.",
			config: []string{"root_import_a_used_in_declare_b"},
			module: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:   "Import passthrough module; instantiate imported declare in a nested declare.",
			config: []string{"root_import_a_used_in_declare_c_within_declare_b"},
			module: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:         "Import passthrough module which also imports a passthrough module; update nested module.",
			config:       []string{"root_import_a"},
			module:       []string{"declare_passthrough_import_a"},
			nestedModule: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "nested_module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:         "Import passthrough module which also imports a passthrough module; update module.",
			config:       []string{"root_import_a"},
			module:       []string{"declare_passthrough_import_a"},
			nestedModule: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:         "Import passthrough module which also imports a passthrough module and uses it inside of a nested declare.",
			config:       []string{"root_import_a"},
			module:       []string{"declare_passthrough_import_a_used_in_declare_b"},
			nestedModule: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "nested_module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:   "Import module with two declares; one used in the other one.",
			config: []string{"root_import_b"},
			module: []string{"declare_passthrough", "declare_instantiates_declare_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_negative_passthrough", "declare_instantiates_declare_passthrough"},
			},
		},
		{
			name:         "Import passthrough module and instantiate it in a declare. The imported module has a nested declare that uses an imported passthrough.",
			config:       []string{"root_import_a_in_declare_b"},
			module:       []string{"declare_passthrough_import_a_used_in_declare_b"},
			nestedModule: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "nested_module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
		{
			name:         "Import passthrough module and update it with an import passthrough",
			config:       []string{"root_import_a"},
			module:       []string{"declare_passthrough"},
			nestedModule: []string{"declare_negative_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_passthrough_import_a"},
			},
		},
		{
			name:         "Import passthrough module which also imports a passthrough module and update it to a simple passthrough",
			config:       []string{"root_import_a"},
			module:       []string{"declare_passthrough_import_a"},
			nestedModule: []string{"declare_passthrough"},
			update: &updateFile{
				name:         "module",
				updateConfig: []string{"declare_negative_passthrough"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer os.Remove("module")
			require.NoError(t, os.WriteFile("module", []byte(concatModules(t, modules, tc.module)), 0664))
			if tc.nestedModule != nil {
				defer os.Remove("nested_module")
				require.NoError(t, os.WriteFile("nested_module", []byte(concatModules(t, modules, tc.nestedModule)), 0664))
			}

			if tc.update != nil {
				testConfig(t, concatModules(t, modules, tc.config), func() {
					require.NoError(t, os.WriteFile(tc.update.name, []byte(concatModules(t, modules, tc.update.updateConfig)), 0664))
				})
			} else {
				testConfig(t, concatModules(t, modules, tc.config), nil)
			}
		})
	}
}

func TestImportError(t *testing.T) {
	modules := loadModules(t)
	testCases := []struct {
		name          string
		config        string
		expectedError string
	}{
		{
			name:          "Imported declare tries to access declare at the root.",
			config:        "import_use_out_of_scope_declare",
			expectedError: `cannot retrieve the definition of component name "cantAccessThis"`,
		},
		{
			name:          "Root tries to access declare in nested import.",
			config:        "root_use_out_of_scope_declare",
			expectedError: `Failed to build component: loading custom component controller: custom component config not found in the registry, namespace: "testImport", componentName: "cantAccessThis"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testConfigError(t, modules[tc.config], tc.expectedError)
		})
	}
}

func TestImportReload(t *testing.T) {
	modules := loadModules(t)
	testCases := []struct {
		name         string
		config       []string
		module       []string
		nestedModule []string
		newConfig    []string
	}{
		{
			name:         "Import passthrough module and instantiate it in a declare. The imported module has a nested declare that uses an imported passthrough.",
			config:       []string{"root_import_a_in_declare_b"},
			module:       []string{"declare_passthrough_import_a_used_in_declare_b"},
			nestedModule: []string{"declare_passthrough"},
			newConfig:    []string{"root_import_a_negative_input"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer os.Remove("module")
			require.NoError(t, os.WriteFile("module", []byte(concatModules(t, modules, tc.module)), 0664))
			if tc.nestedModule != nil {
				defer os.Remove("nested_module")
				require.NoError(t, os.WriteFile("nested_module", []byte(concatModules(t, modules, tc.nestedModule)), 0664))
			}
			testConfigReload(t, concatModules(t, modules, tc.config), concatModules(t, modules, tc.newConfig))
		})
	}
}

func TestImportString(t *testing.T) {
	modules := loadModules(t)
	t.Run("Import a declare and instantiate a declare from it", func(t *testing.T) {
		defer verifyNoGoroutineLeaks(t)
		testConfig(t, modules["import_string_declare_passthrough"], nil)
	})
}

func testConfig(t *testing.T, config string, update func()) {
	defer verifyNoGoroutineLeaks(t)
	ctrl := flow.New(testOptions(t))
	f, err := flow.ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.Run(ctx)
	}()

	// Check for initial condition
	require.Eventually(t, func() bool {
		export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
		return export.LastAdded >= 10
	}, 3*time.Second, 10*time.Millisecond)

	if update != nil {
		update()

		// Export should be -10 after update
		require.Eventually(t, func() bool {
			export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
			return export.LastAdded <= -10
		}, 3*time.Second, 10*time.Millisecond)
	}
}

func testConfigReload(t *testing.T, config string, newConfig string) {
	defer verifyNoGoroutineLeaks(t)
	ctrl := flow.New(testOptions(t))
	f, err := flow.ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.Run(ctx)
	}()

	// Check for initial condition
	require.Eventually(t, func() bool {
		export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
		return export.LastAdded >= 10
	}, 3*time.Second, 10*time.Millisecond)

	f, err = flow.ParseSource(t.Name(), []byte(newConfig))
	require.NoError(t, err)
	require.NotNil(t, f)

	// Reload the controller with the new config.
	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)

	// Export should be -10 after update
	require.Eventually(t, func() bool {
		export := getExport[testcomponents.SummationExports](t, ctrl, "", "testcomponents.summation.sum")
		return export.LastAdded <= -10
	}, 3*time.Second, 10*time.Millisecond)
}

func testConfigError(t *testing.T, config string, expectedError string) {
	defer verifyNoGoroutineLeaks(t)
	s, err := logging.New(os.Stderr, logging.DefaultOptions)
	require.NoError(t, err)
	ctrl := flow.New(flow.Options{
		Logger:   s,
		DataPath: t.TempDir(),
		Reg:      nil,
		Services: []service.Service{},
	})
	f, err := flow.ParseSource(t.Name(), []byte(config))
	require.NoError(t, err)
	require.NotNil(t, f)

	err = ctrl.LoadSource(f, nil)
	require.ErrorContains(t, err, expectedError)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.Run(ctx)
	}()
}

// loadModules load river configs from a txtar file to a map
func loadModules(t *testing.T) map[string]string {
	archive, err := txtar.ParseFile("import_test.txtar")
	require.NoError(t, err)

	modules := make(map[string]string)
	for _, file := range archive.Files {
		modules[file.Name] = string(file.Data)
	}
	return modules
}

// concatModules concatenates modules into one module
func concatModules(t *testing.T, modules map[string]string, files []string) string {
	m := make([]string, len(files))
	for i, f := range files {
		mod, found := modules[f]
		require.True(t, found, f)
		m[i] = mod
	}
	return strings.Join(m, "\n")
}
