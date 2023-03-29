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
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/flow/logging"
	"github.com/stretchr/testify/require"
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
		argument "username" {
		} 
		argument "password" {
		}`

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
    arguments = {
        username = module.string.exporter.exports.username,
        password = module.string.exporter.exports.password,
    }
}`
	fmtFile := strings.Replace(riverFile, "%exp", exportPath, 1)
	fmtFile = strings.Replace(fmtFile, "%arg", argPath, 1)

	f := flow.New(testOptions(t))
	ff, err := flow.ReadFile("test", []byte(fmtFile))
	require.NoError(t, err)
	err = f.LoadFile(ff, nil)
	require.NoError(t, err)
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, 10*time.Second)
	defer cncl()
	go f.Run(ctx)
	time.Sleep(100 * time.Millisecond)
	infos := f.ComponentInfos()
	for _, i := range infos {
		if i.ID != "module.string.importer" {
			continue
		}
		buf := bytes.NewBuffer(nil)
		err = f.ComponentJSON(buf, i)
		require.NoError(t, err)
		require.True(t, strings.Contains(buf.String(), "password1"))
		require.True(t, strings.Contains(buf.String(), "bob"))
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

	return flow.Options{
		LogSink:  s,
		DataPath: t.TempDir(),
		Reg:      nil,
	}
}
