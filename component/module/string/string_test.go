//go:build linux

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
	"github.com/grafana/agent/pkg/module"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestExports(t *testing.T) {
	exportContent := `
		export "username" {
			value = "bob"
		}
		export "password" {
			value = "password1"
		}`
	argumentContent := `
		argument "username" {} 
		argument "password" {}`

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.river")
	argPath := filepath.Join(tmpDir, "args.river")
	writeFile(t, exportPath, exportContent)
	writeFile(t, argPath, argumentContent)

	riverFile := `
		local.file "exporter" { filename = "%exp" }
		local.file "args"     { filename = "%arg" }
		
		module.string "exporter" {
			content = local.file.exporter.content
		}
		
		module.string "importer" {
			content = local.file.args.content
			arguments {
				username = module.string.exporter.exports.username
				password = module.string.exporter.exports.password
			}
		}`
	fmtFile := strings.Replace(riverFile, "%exp", exportPath, 1)
	fmtFile = strings.Replace(fmtFile, "%arg", argPath, 1)
	testFile(t, fmtFile, "module.string.importer", []string{"password1", "bob"})
}

func TestUpdatingExports(t *testing.T) {
	// The tick ensures that the dummy value will get exported multiple times.
	// In previous versions this would cause ONLY the dummy value to be passed to exports.
	// When in fact they all should be kept.
	exportContent := `
		testcomponents.tick "t1" {
		  frequency = "1s"
		}
		export "address" {
		  value = "localhost:12345"
		}
		
		export "username" {
		  value = "23766"
		}
		
		export "dummy" {
		  value = testcomponents.tick.t1.tick_time
		}`

	loaderContent := `
		local.file "load_export" {
			filename = "%exports%"
		}
		
		module.string "loadexport" {
			content = local.file.load_export.content
		}
		
		module.string "easy_load" {
			content = ""
			arguments {
				address = module.string.loadexport.exports.address
				username = module.string.loadexport.exports.username
			}
		}`

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.river")
	writeFile(t, exportPath, exportContent)
	fmtFile := strings.Replace(loaderContent, "%exports%", exportPath, 1)
	testFile(t, fmtFile, "module.string.loadexport", []string{"address", "dummy"})
}

func testFile(t *testing.T, fmtFile string, componentToFind string, searchable []string) {
	f := flow.New(testOptions(t), "")
	ff, err := flow.ReadFile("test", []byte(fmtFile))
	require.NoError(t, err)
	err = f.LoadFile(ff, nil)
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, 10*time.Second)
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

	l := util.TestFlowLogger(t)
	c := &cluster.Clusterer{Node: cluster.NewLocalNode("")}
	return flow.Options{
		Logger:    l,
		DataPath:  t.TempDir(),
		Reg:       nil,
		Clusterer: c,
		Controller: module.NewModule(&module.Options{
			Logger:    l,
			Tracer:    trace.NewNoopTracerProvider(),
			Clusterer: c,
			Reg:       prometheus.DefaultRegisterer,
		}),
	}
}
