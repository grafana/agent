package prometheusconvert_test

import (
	"testing"

	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/converter/internal/test_common"
)

func TestConvert(t *testing.T) {
	test_common.TestDirectory(t, "testdata", ".yaml", prometheusconvert.Convert)
}
