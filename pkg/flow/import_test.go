package flow_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
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

// The tests are using the .txtar files stored in the testdata folder.

type testImportFile struct {
	description       string      // description at the top of the txtar file
	main              string      // root config that the controller should load
	module            string      // module imported by the root config
	nestedModule      string      // nested module that can be imported by the module
	reloadConfig      string      // root config that the controller should apply on reload
	otherNestedModule string      // another nested module
	update            *updateFile // update can be used to update the content of a file at runtime
}

type updateFile struct {
	name         string // name of the file which should be updated
	updateConfig string // new module config which should be used
}

func buildTestImportFile(t *testing.T, filename string) testImportFile {
	archive, err := txtar.ParseFile(filename)
	require.NoError(t, err)
	var tc testImportFile
	tc.description = string(archive.Comment)
	for _, riverConfig := range archive.Files {
		switch riverConfig.Name {
		case "main.river":
			tc.main = string(riverConfig.Data)
		case "module.river":
			tc.module = string(riverConfig.Data)
		case "nested_module.river":
			tc.nestedModule = string(riverConfig.Data)
		case "update/module.river":
			require.Nil(t, tc.update)
			tc.update = &updateFile{
				name:         "module.river",
				updateConfig: string(riverConfig.Data),
			}
		case "update/nested_module.river":
			require.Nil(t, tc.update)
			tc.update = &updateFile{
				name:         "nested_module.river",
				updateConfig: string(riverConfig.Data),
			}
		case "reload_config.river":
			tc.reloadConfig = string(riverConfig.Data)
		case "other_nested_module.river":
			tc.otherNestedModule = string(riverConfig.Data)
		}
	}
	return tc
}

func TestImportFile(t *testing.T) {
	directory := "./testdata/import_file"
	for _, file := range getTestFiles(directory, t) {
		tc := buildTestImportFile(t, filepath.Join(directory, file.Name()))
		t.Run(tc.description, func(t *testing.T) {
			defer os.Remove("module.river")
			require.NoError(t, os.WriteFile("module.river", []byte(tc.module), 0664))
			if tc.nestedModule != "" {
				defer os.Remove("nested_module.river")
				require.NoError(t, os.WriteFile("nested_module.river", []byte(tc.nestedModule), 0664))
			}
			if tc.otherNestedModule != "" {
				defer os.Remove("other_nested_module.river")
				require.NoError(t, os.WriteFile("other_nested_module.river", []byte(tc.otherNestedModule), 0664))
			}

			if tc.update != nil {
				testConfig(t, tc.main, tc.reloadConfig, func() {
					require.NoError(t, os.WriteFile(tc.update.name, []byte(tc.update.updateConfig), 0664))
				})
			} else {
				testConfig(t, tc.main, tc.reloadConfig, nil)
			}
		})
	}
}

func TestImportString(t *testing.T) {
	directory := "./testdata/import_string"
	for _, file := range getTestFiles(directory, t) {
		archive, err := txtar.ParseFile(filepath.Join(directory, file.Name()))
		require.NoError(t, err)
		t.Run(archive.Files[0].Name, func(t *testing.T) {
			testConfig(t, string(archive.Files[0].Data), "", nil)
		})
	}
}

func TestImportGit(t *testing.T) {
	directory := "./testdata/import_git"
	for _, file := range getTestFiles(directory, t) {
		archive, err := txtar.ParseFile(filepath.Join(directory, file.Name()))
		require.NoError(t, err)
		t.Run(archive.Files[0].Name, func(t *testing.T) {
			testConfig(t, string(archive.Files[0].Data), "", nil)
		})
	}
}

func TestImportHTTP(t *testing.T) {
	directory := "./testdata/import_http"
	for _, file := range getTestFiles(directory, t) {
		archive, err := txtar.ParseFile(filepath.Join(directory, file.Name()))
		require.NoError(t, err)
		t.Run(archive.Files[0].Name, func(t *testing.T) {
			testConfig(t, string(archive.Files[0].Data), "", nil)
		})
	}
}

type testImportError struct {
	description   string
	main          string
	expectedError string
}

func buildTestImportError(t *testing.T, filename string) testImportError {
	archive, err := txtar.ParseFile(filename)
	require.NoError(t, err)
	var tc testImportError
	tc.description = string(archive.Comment)
	for _, riverConfig := range archive.Files {
		switch riverConfig.Name {
		case "main.river":
			tc.main = string(riverConfig.Data)
		case "error":
			tc.expectedError = string(riverConfig.Data)
		}
	}
	return tc
}

func TestImportError(t *testing.T) {
	directory := "./testdata/import_error"
	for _, file := range getTestFiles(directory, t) {
		tc := buildTestImportError(t, filepath.Join(directory, file.Name()))
		t.Run(tc.description, func(t *testing.T) {
			testConfigError(t, tc.main, strings.TrimRight(tc.expectedError, "\n"))
		})
	}
}

func TestPullUpdating(t *testing.T) {
	// Previously we used fetch instead of pull, which would set the FETCH_HEAD but not HEAD
	// This caused changes not to propagate if there were changes, since HEAD was pinned to whatever it was on the initial download.
	// Switching to pull removes this problem at the expense of network bandwidth.
	// Tried switching to FETCH_HEAD but FETCH_HEAD is only set on fetch and not initial repo clone so we would need to
	// remember to always call fetch after clone.
	//
	// This test ensures we can pull the correct values down if they update no matter what, it works by creating a local
	// file based git repo then committing a file, running the component, then updating the file in the repo.
	testRepo := t.TempDir()

	contents := `declare "add" {
    argument "a" {}
    argument "b" {}

    export "sum" {
        value = argument.a.value + argument.b.value
    }
}`
	main := `
import.git "testImport" {
	repository = "` + testRepo + `"
  	path = "math.river"
    pull_frequency = "5s"
}

testImport.add "cc" {
	a = 1
    b = 1
}
`
	init := exec.Command("git", "init", testRepo)
	err := init.Run()
	require.NoError(t, err)
	math := filepath.Join(testRepo, "math.river")
	err = os.WriteFile(math, []byte(contents), 0666)
	require.NoError(t, err)
	add := exec.Command("git", "add", ".")
	add.Dir = testRepo
	err = add.Run()
	require.NoError(t, err)
	commit := exec.Command("git", "commit", "-m \"test\"")
	commit.Dir = testRepo
	err = commit.Run()
	require.NoError(t, err)

	defer verifyNoGoroutineLeaks(t)
	ctrl, f := setup(t, main)
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
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 2
	}, 3*time.Second, 10*time.Millisecond)

	contentsMore := `declare "add" {
    argument "a" {}
    argument "b" {}

    export "sum" {
        value = argument.a.value + argument.b.value + 1
    }
}`
	err = os.WriteFile(math, []byte(contentsMore), 0666)
	require.NoError(t, err)
	add2 := exec.Command("git", "add", ".")
	add2.Dir = testRepo
	add2.Run()

	commit2 := exec.Command("git", "commit", "-m \"test2\"")
	commit2.Dir = testRepo
	commit2.Run()

	// Check for final condition.
	require.Eventually(t, func() bool {
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 3
	}, 20*time.Second, 1*time.Millisecond)
}

func testConfig(t *testing.T, config string, reloadConfig string, update func()) {
	defer verifyNoGoroutineLeaks(t)
	ctrl, f := setup(t, config)

	err := ctrl.LoadSource(f, nil)
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

	if reloadConfig != "" {
		f, err = flow.ParseSource(t.Name(), []byte(reloadConfig))
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
}

func testConfigError(t *testing.T, config string, expectedError string) {
	defer verifyNoGoroutineLeaks(t)
	ctrl, f := setup(t, config)
	err := ctrl.LoadSource(f, nil)
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

func setup(t *testing.T, config string) (*flow.Flow, *flow.Source) {
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
	return ctrl, f
}

func getTestFiles(directory string, t *testing.T) []fs.FileInfo {
	dir, err := os.Open(directory)
	require.NoError(t, err)
	defer dir.Close()

	files, err := dir.Readdir(-1)
	require.NoError(t, err)

	return files
}
