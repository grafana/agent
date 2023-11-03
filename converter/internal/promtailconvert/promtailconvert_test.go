package promtailconvert_test

import (
	"testing"

	"github.com/grafana/agent/converter/internal/promtailconvert"
	"github.com/grafana/agent/converter/internal/test_common"
	_ "github.com/grafana/agent/pkg/metrics/instance" // Imported to override default values via the init function.
)

func TestConvert(t *testing.T) {
	test_common.TestDirectory(t, "testdata", ".yaml", true, []string{}, promtailconvert.Convert)
}
