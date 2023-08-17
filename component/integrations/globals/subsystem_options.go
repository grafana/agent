package globals

import (
	internal "github.com/grafana/agent/pkg/integrations/v2"
	"github.com/grafana/agent/pkg/integrations/v2/autoscrape"
)

type subsystemOptions struct {
	metrics metricsSubsystemOptions `river:"metrics,block,optional`

	configs internal.Configs `river:""`
}

type metricsSubsystemOptions struct {
	autoscrape autoscrape.Global `river:"autoscrape,attr,optional"`
}
