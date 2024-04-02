package promtailconvert_test

import (
	"testing"

	"github.com/grafana/agent/internal/converter/internal/promtailconvert"
	"github.com/grafana/agent/internal/converter/internal/test_common"
	_ "github.com/grafana/agent/internal/static/metrics/instance" // Imported to override default values via the init function.
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	test_common.TestDirectory(t, "testdata", ".yaml", true, []string{}, promtailconvert.Convert)
}

// TestGlobalPositions will ensure that legacy is set if there are multiple readers. Since the global positions file contains
// all files but when converted only the files of the actual individual jobs should be tracked. This does lead to a case
// where there are orphaned files but since they will not be read by the job they are noops. If they add the previously unused
// filepath to a job, say for instance add the fun paths to fun2 then it will pick up where it left off from the conversion
// but this is likely the best path and unlikely to be a common occurrence.
func TestGlobalPositions(t *testing.T) {
	test := `
positions:
  filename: /good/positions.yml
scrape_configs:
  - job_name: fun
    file_sd_configs:
      - files:
          - /etc/prometheus/targets/*.json
        refresh_interval: 5m
      - files:
          - /etc/agent/targets/*.json
        refresh_interval: 30m
  - job_name: fun2
    file_sd_configs:
      - files:
          - /etc/prometheus/targets2/*.json
        refresh_interval: 5m
      - files:
          - /etc/agent/targets2/*.json
        refresh_interval: 30m
tracing: {enabled: false}
server: {register_instrumentation: false}
`
	expected := `discovery.file "fun" {
	files = ["/etc/prometheus/targets/*.json"]
}

discovery.file "fun_2" {
	files            = ["/etc/agent/targets/*.json"]
	refresh_interval = "30m0s"
}

local.file_match "fun" {
	path_targets = concat(
		discovery.file.fun.targets,
		discovery.file.fun_2.targets,
	)
}

loki.source.file "fun" {
	targets               = local.file_match.fun.targets
	forward_to            = []
	legacy_positions_file = "/good/positions.yml"
}

discovery.file "fun2" {
	files = ["/etc/prometheus/targets2/*.json"]
}

discovery.file "fun2_2" {
	files            = ["/etc/agent/targets2/*.json"]
	refresh_interval = "30m0s"
}

local.file_match "fun2" {
	path_targets = concat(
		discovery.file.fun2.targets,
		discovery.file.fun2_2.targets,
	)
}

loki.source.file "fun2" {
	targets               = local.file_match.fun2.targets
	forward_to            = []
	legacy_positions_file = "/good/positions.yml"
}
`
	out, diags := promtailconvert.Convert([]byte(test), []string{})
	require.True(t, diags.Error() == "")
	require.NotNil(t, out)
	test_common.ValidateRiver(t, []byte(expected), out, false)
}
