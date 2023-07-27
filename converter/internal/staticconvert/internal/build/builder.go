package build

import (
	"github.com/grafana/agent/converter/diag"
	"github.com/grafana/agent/pkg/config"
	"github.com/grafana/agent/pkg/integrations/node_exporter"
	"github.com/grafana/agent/pkg/river/token/builder"
)

type IntegrationsV1ConfigBuilder struct {
	f         *builder.File
	diags     *diag.Diagnostics
	cfg       *config.Config
	globalCtx *GlobalContext
}

func NewIntegrationsV1ConfigBuilder(f *builder.File, diags *diag.Diagnostics, cfg *config.Config, globalCtx *GlobalContext) *IntegrationsV1ConfigBuilder {
	return &IntegrationsV1ConfigBuilder{
		f:         f,
		diags:     diags,
		cfg:       cfg,
		globalCtx: globalCtx,
	}
}

func (b *IntegrationsV1ConfigBuilder) AppendIntegrations() {
	for _, integration := range b.cfg.Integrations.ConfigV1.Integrations {
		switch itg := integration.Config.(type) {
		case *node_exporter.Config:
			b.AppendNodeExporter(itg)
		}
	}
}
