package staticconvert_test

import (
	"runtime"
	"testing"

	"github.com/grafana/agent/converter/internal/staticconvert"
	"github.com/grafana/agent/converter/internal/test_common"
)

func TestConvert(t *testing.T) {
	test_common.TestDirectory(t, "testdata", ".yaml", staticconvert.Convert)

	if runtime.GOOS == "windows" {
		test_common.TestDirectory(t, "testdata_windows", ".yaml", staticconvert.Convert)
	}
}
