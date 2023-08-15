package mongodb

import (
	"net/url"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/mongodb_exporter"
	"github.com/grafana/agent/pkg/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.mongodb",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: exporter.RequiredServices(),
		Build:         exporter.NewWithTargetBuilder(createExporter, "mongodb", customizeTarget),
	})
}

func createExporter(opts component.Options, args component.Arguments) (integrations.Integration, error) {
	a := args.(Arguments)
	return a.Convert().NewIntegration(opts.Logger)
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget

	u, _ := url.Parse(string(a.URI))
	target["instance"] = u.Host
	return []discovery.Target{target}
}

type Arguments struct {
	URI                    rivertypes.Secret `river:"mongodb_uri,attr"`
	DirectConnect          bool              `river:"direct_connect,attr,optional"`
	DiscoveringMode        bool              `river:"discovering_mode,attr,optional"`
	TLSBasicAuthConfigPath string            `river:"tls_basic_auth_config_path,attr,optional"`
}

func (a *Arguments) Convert() *mongodb_exporter.Config {
	return &mongodb_exporter.Config{
		URI:                    config_util.Secret(a.URI),
		DirectConnect:          a.DirectConnect,
		DiscoveringMode:        a.DiscoveringMode,
		TLSBasicAuthConfigPath: a.TLSBasicAuthConfigPath,
	}
}
