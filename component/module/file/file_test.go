package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/flow"
	"github.com/grafana/agent/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestModule(t *testing.T) {
	tempDir := t.TempDir()
	debugLevelFilePath := filepath.Join(tempDir, "debug_level.txt")
	os.WriteFile(debugLevelFilePath, []byte("info"), 0664)

	tt := []struct {
		name                                   string
		moduleContents                         string
		expectedHealthType                     component.HealthType
		expectedHealthMessagePrefix            string
		expectedManagedFileHealthType          component.HealthType
		expectedManagedFileHealthMessagePrefix string
	}{
		{
			name: "Good Module",

			moduleContents: `local.file "log_level" {
				filename  = "` + riverEscape(debugLevelFilePath) + `"
			}`,
			expectedHealthType:          component.HealthTypeHealthy,
			expectedHealthMessagePrefix: "started component",

			expectedManagedFileHealthType:          component.HealthTypeHealthy,
			expectedManagedFileHealthMessagePrefix: "read file",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := flow.New(testOptions(t))

			moduleFilePath := filepath.Join(tempDir, "module.river")
			os.WriteFile(moduleFilePath, []byte(tc.moduleContents), 0664)

			flowFile := fmt.Sprintf(`
            module.file "test" {
				filename = "%s"
			}
			`, riverEscape(moduleFilePath))

			ff, err := flow.ReadFile("test", []byte(flowFile))
			require.NoError(t, err)
			err = f.LoadFile(ff, nil)
			require.NoError(t, err)
			ctx := context.Background()
			ctx, cncl := context.WithTimeout(ctx, 10*time.Second)
			defer cncl()
			go f.Run(ctx)

			require.Eventually(
				t,
				func() bool {
					infos := f.ComponentInfos()
					for _, i := range infos {
						if i.ID != "module.file.test" {
							continue
						}
						return i.Health.State == tc.expectedHealthType.String() && strings.HasPrefix(tc.expectedHealthMessagePrefix, i.Health.Message)
					}
					return false
				},
				5*time.Second,
				50*time.Millisecond,
				"did not reach required health status before timeout: %v",
				tc.expectedHealthType,
			)
		})
	}
}

func TestBadFile(t *testing.T) {
	tt := []struct {
		name                  string
		moduleContents        string
		expectedErrorContains string
	}{
		{
			name:                  "Bad Module",
			moduleContents:        `this isn't a valid module config`,
			expectedErrorContains: `expected block label, got IDENT`,
		},
		{
			name:                  "Bad Component",
			moduleContents:        `local.fake "fake" {}`,
			expectedErrorContains: `Unrecognized component name "local.fake"`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := flow.New(testOptions(t))
			tempDir := t.TempDir()
			moduleFilePath := filepath.Join(tempDir, "module.river")
			os.WriteFile(moduleFilePath, []byte(tc.moduleContents), 0664)

			flowFile := fmt.Sprintf(`
            module.file "test" {
				filename = "%s"
			}
			`, riverEscape(moduleFilePath))

			ff, err := flow.ReadFile("test", []byte(flowFile))
			require.NoError(t, err)
			err = f.LoadFile(ff, nil)

			require.ErrorContains(t, err, tc.expectedErrorContains)
		})
	}
}

func riverEscape(filePath string) string {
	if runtime.GOOS == "windows" {
		return strings.Replace(filePath, `\`, `\\`, -1)
	}

	return filePath
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
	}
}
