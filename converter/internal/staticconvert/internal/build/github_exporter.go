package build

import (
	"fmt"

	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter/github"
	"github.com/grafana/agent/converter/internal/common"
	"github.com/grafana/agent/converter/internal/prometheusconvert"
	"github.com/grafana/agent/pkg/integrations/github_exporter"
	"github.com/grafana/river/rivertypes"
)

func (b *IntegrationsV1ConfigBuilder) appendGithubExporter(config *github_exporter.Config) discovery.Exports {
	args := toGithubExporter(config)
	compLabel := common.LabelForParts(b.globalCtx.LabelPrefix, config.Name())
	b.f.Body().AppendBlock(common.NewBlockWithOverride(
		[]string{"prometheus", "exporter", "github"},
		compLabel,
		args,
	))

	return prometheusconvert.NewDiscoveryExports(fmt.Sprintf("prometheus.exporter.github.%s.targets", compLabel))
}

func toGithubExporter(config *github_exporter.Config) *github.Arguments {
	return &github.Arguments{
		APIURL:        config.APIURL,
		Repositories:  config.Repositories,
		Organizations: config.Organizations,
		Users:         config.Users,
		APIToken:      rivertypes.Secret(config.APIToken),
		APITokenFile:  config.APITokenFile,
	}
}
