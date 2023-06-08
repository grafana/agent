//go:build !windows

package string

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/grafana/agent/component/local/file"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
)

const loggingConfig = `
	logging {}`

const tracingConfig = `
	tracing {}`

const argumentConfig = `
	argument "username" {} 
	argument "defaulted" {
		optional = true
		default = "default_value"
	}`

const argumentModuleLoaderConfig = `
	local.file "args"     { filename = "%arg" }
	module.string "importer" {
		content = local.file.args.content
		arguments {
			username = module.string.exporter.exports.username
		}
	}`

const exportStringConfig = `
	export "username" {
		value = "bob"
	}`

const exportComponentConfig = `
	testcomponents.tick "t1" {
		frequency = "1s"
  	}

	export "dummy" {
		value = testcomponents.tick.t1.tick_time
	}`

const exportModuleLoaderConfig = `
	local.file "exporter" { filename = "%exp" }
	
	module.string "exporter" {
		content = local.file.exporter.content
	}`

func TestModule(t *testing.T) {
	tt := []struct {
		name                  string
		riverContent          string
		argumentModuleContent string
		exportModuleContent   string
		expectedComponentId   string
		expectedExports       []string
		expectedErrorContains string
	}{
		{
			name:                  "Export String",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedComponentId:   "module.string.importer",
			expectedExports:       []string{"bob"},
		},
		{
			name:                  "Export Component",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig + exportComponentConfig,
			expectedComponentId:   "module.string.exporter",
			expectedExports:       []string{"username", "dummy"},
		},
		{
			name:         "Empty Content Allowed",
			riverContent: `module.string "empty" { content = "" }`,
		},
		{
			name:                  "Argument blocks not allowed in parent config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig + argumentConfig,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "argument blocks only allowed inside a module",
		},
		{
			name:                  "Export blocks not allowed in parent config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig + exportStringConfig,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "export blocks only allowed inside a module",
		},
		{
			name:                  "Logging blocks not allowed in module config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig,
			argumentModuleContent: argumentConfig + loggingConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "logging block not allowed inside a module",
		},
		{
			name:                  "Tracing blocks not allowed in module config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig,
			argumentModuleContent: argumentConfig + tracingConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "tracing block not allowed inside a module",
		},
		{
			name:                  "Argument not defined in module source",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig,
			argumentModuleContent: `argument "different_argument" {}`,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "Provided argument \"username\" is not defined in the module",
		},
		{
			name: "Missing required argument",
			riverContent: exportModuleLoaderConfig + `
				local.file "args" { filename = "%arg" }
				module.string "importer" {
					content = local.file.args.content
				}`,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "Failed to evaluate node for config block: missing required argument \"username\" to module",
		},
		{
			name:                  "Duplicate logging config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig + loggingConfig + loggingConfig,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "\"logging\" block already declared",
		},
		{
			name:                  "Duplicate tracing config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig + tracingConfig + tracingConfig,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "\"tracing\" block already declared",
		},
		{
			name:                  "Duplicate argument config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig,
			argumentModuleContent: argumentConfig + argumentConfig,
			exportModuleContent:   exportStringConfig,
			expectedErrorContains: "\"argument.username\" block already declared",
		},
		{
			name:                  "Duplicate export config",
			riverContent:          argumentModuleLoaderConfig + exportModuleLoaderConfig,
			argumentModuleContent: argumentConfig,
			exportModuleContent:   exportStringConfig + exportStringConfig,
			expectedErrorContains: "\"export.username\" block already declared",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			riverFile := tc.riverContent

			// Prep the argument module file
			if tc.argumentModuleContent != "" {
				argPath := filepath.Join(tmpDir, "args.river")
				writeFile(t, argPath, tc.argumentModuleContent)
				riverFile = strings.Replace(riverFile, "%arg", argPath, 1)
			}

			// Prep the export module file
			if tc.exportModuleContent != "" {
				exportPath := filepath.Join(tmpDir, "export.river")
				writeFile(t, exportPath, tc.exportModuleContent)
				riverFile = strings.Replace(riverFile, "%exp", exportPath, 1)
			}

			testFile(t, riverFile, tc.expectedComponentId, tc.expectedExports, tc.expectedErrorContains)
		})
	}
}

func testFile(t *testing.T, fmtFile string, componentToFind string, searchable []string, expectedErrorContains string) {
	f := flow.New(testOptions(t))
	ff, err := flow.ReadFile("test", []byte(fmtFile))
	require.NoError(t, err)
	err = f.LoadFile(ff, nil)
	if expectedErrorContains == "" {
		require.NoError(t, err)
	} else {
		require.ErrorContains(t, err, expectedErrorContains)
		return
	}

	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, 20*time.Second)
	defer cncl()
	go f.Run(ctx)
	time.Sleep(3 * time.Second)
	infos := f.ComponentInfos()
	for _, i := range infos {
		if i.ID != componentToFind {
			continue
		}
		buf := bytes.NewBuffer(nil)
		err = f.ComponentJSON(buf, i)
		require.NoError(t, err)
		// This ensures that although dummy is not used it still exists.
		// And that multiple exports are displayed when only one of them is updating.
		for _, s := range searchable {
			require.True(t, strings.Contains(buf.String(), s))
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	err := os.WriteFile(path, []byte(content), 0666)
	require.NoError(t, err)
}

func testOptions(t *testing.T) flow.Options {
	t.Helper()

	s, err := logging.WriterSink(os.Stderr, logging.DefaultSinkOptions)
	require.NoError(t, err)

	c := &cluster.Clusterer{Node: cluster.NewLocalNode("")}

	return flow.Options{
		LogSink:   s,
		DataPath:  t.TempDir(),
		Reg:       nil,
		Clusterer: c,
	}
}
