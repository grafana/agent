package staticconvert_test

import (
	"runtime"
	"testing"

	"github.com/grafana/agent/converter/internal/staticconvert"
	"github.com/grafana/agent/converter/internal/test_common"
	_ "github.com/grafana/agent/pkg/metrics/instance" // Imported to override default values via the init function.
)

func TestConvert(t *testing.T) {
	test_common.TestDirectory(t, "testdata", ".yaml", staticconvert.Convert)
	test_common.TestDirectory(t, "testdata2", ".yaml", staticconvert.Convert)

	if runtime.GOOS == "windows" {
		test_common.TestDirectory(t, "testdata_windows", ".yaml", staticconvert.Convert)
	}
}
