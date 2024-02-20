package flowtest_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/grafana/agent/pkg/flow/flowtest"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	filepath.WalkDir("testdata/", func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) != ".txtar" {
			return nil
		}
		t.Run(d.Name(), func(t *testing.T) {
			require.NoError(t, flowtest.TestScript(path))
		})
		return nil
	})
}
