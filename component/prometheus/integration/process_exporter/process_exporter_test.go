package process_exporter

import (
	"testing"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

func TestRiverConfig(t *testing.T) {
	var exampleRiverConfig = `
	process_names {
		comm    = ["process_name_1", "grafana-agent"]
		exe     = ["/usr/bin/process"]
		cmdline = ["*flow*"]
	}
	track_children    = false
	track_threads     = false
	gather_smaps      = true
	recheck_on_scrape = true
`

	var args Arguments
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)
}
