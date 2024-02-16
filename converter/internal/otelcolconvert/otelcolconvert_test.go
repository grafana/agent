package otelcolconvert_test

import (
	"testing"

	"github.com/grafana/agent/converter/internal/otelcolconvert"
	"github.com/grafana/agent/converter/internal/test_common"
)

func TestConvert(t *testing.T) {
	// TODO(rfratto): support -update flag.
	test_common.TestDirectory(t, "testdata", ".yaml", true, []string{}, otelcolconvert.Convert)
}
