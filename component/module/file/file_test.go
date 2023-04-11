package file

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/cluster"
	"github.com/grafana/agent/pkg/river"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
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
			expectedHealthMessagePrefix: "module content loaded",

			expectedManagedFileHealthType:          component.HealthTypeHealthy,
			expectedManagedFileHealthMessagePrefix: "read file",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			moduleFilePath := filepath.Join(tempDir, "module.river")
			os.WriteFile(moduleFilePath, []byte(tc.moduleContents), 0664)

			opts := component.Options{
				ID:            "module.file.test",
				Logger:        util.TestFlowLogger(t),
				Clusterer:     noOpClusterer(),
				Registerer:    prometheus.NewRegistry(),
				OnStateChange: func(e component.Exports) {},
				DataPath:      t.TempDir(),
			}

			moduleFileConfig := `filename = "` + riverEscape(moduleFilePath) + `"`

			var args Arguments
			require.NoError(t, river.Unmarshal([]byte(moduleFileConfig), &args))

			c, err := New(opts, args)
			require.NoError(t, err)

			go c.Run(context.Background())
			require.Eventually(
				t,
				func() bool { return tc.expectedHealthType == c.CurrentHealth().Health },
				5*time.Second,
				50*time.Millisecond,
				"did not reach required health status before timeout: %v != %v",
				tc.expectedHealthType,
				c.CurrentHealth().Health,
			)

			require.Equal(t, tc.expectedHealthType, c.CurrentHealth().Health)
			requirePrefix(t, c.CurrentHealth().Message, tc.expectedHealthMessagePrefix)

			require.Equal(t, tc.expectedManagedFileHealthType, c.managedLocalFile.CurrentHealth().Health)
			requirePrefix(t, c.managedLocalFile.CurrentHealth().Message, tc.expectedManagedFileHealthMessagePrefix)
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
		{
			name:                  "Missing Module",
			moduleContents:        "",
			expectedErrorContains: `failed to read file:`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			moduleFilePath := filepath.Join(t.TempDir(), "module.river")
			if tc.moduleContents != "" {
				os.WriteFile(moduleFilePath, []byte(tc.moduleContents), 0664)
			}

			moduleFileConfig := `filename = "` + riverEscape(moduleFilePath) + `"`

			var args Arguments
			require.NoError(t, river.Unmarshal([]byte(moduleFileConfig), &args))

			opts := component.Options{
				ID:            "module.file.test",
				Logger:        util.TestFlowLogger(t),
				Clusterer:     noOpClusterer(),
				Registerer:    prometheus.NewRegistry(),
				OnStateChange: func(e component.Exports) {},
				DataPath:      t.TempDir(),
			}

			_, err := New(opts, args)
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

func requirePrefix(t *testing.T, s string, prefix string) {
	require.True(
		t,
		strings.HasPrefix(s, prefix),
		"expected '%v' to have '%v' prefix",
		s,
		prefix,
	)
}

func noOpClusterer() *cluster.Clusterer {
	return &cluster.Clusterer{Node: cluster.NewLocalNode("")}
}
