package vsphere

import (
	"fmt"
	"net/url"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"github.com/grafana/agent/component/prometheus/exporter"
	"github.com/grafana/agent/pkg/integrations"
	"github.com/grafana/agent/pkg/integrations/vmware_exporter"
	"github.com/grafana/river/rivertypes"
	config_util "github.com/prometheus/common/config"
)

func init() {
	component.Register(component.Registration{
		Name:          "prometheus.exporter.vsphere",
		Args:          Arguments{},
		Exports:       exporter.Exports{},
		NeedsServices: exporter.RequiredServices(),
		Build:         exporter.NewWithTargetBuilder(createExporter, "vsphere", customizeTarget),
	})
}

func createExporter(opts component.Options, args component.Arguments, defaultInstanceKey string) (integrations.Integration, string, error) {
	a := args.(Arguments)
	return integrations.NewIntegrationWithInstanceKey(opts.Logger, a.Convert(), defaultInstanceKey)
}

func customizeTarget(baseTarget discovery.Target, args component.Arguments) []discovery.Target {
	a := args.(Arguments)
	target := baseTarget

	url, err := url.Parse(a.VSphereURL)
	if err != nil {
		return []discovery.Target{target}
	}

	target["instance"] = fmt.Sprintf("%s:%s", url.Hostname(), url.Port())
	return []discovery.Target{target}
}

// DefaultArguments holds the default settings for the vsphere exporter
var DefaultArguments = Arguments{
	ChunkSize:               256,
	CollectConcurrency:      8,
	ObjectDiscoveryInterval: 0,
	EnableExporterMetrics:   true,
}

// Arguments controls the vsphere exporter.
type Arguments struct {
	ChunkSize               int               `river:"request_chunk_size,attr,optional"`
	CollectConcurrency      int               `river:"collect_concurrency,attr,optional"`
	VSphereURL              string            `river:"vsphere_url,attr,optional"`
	VSphereUser             string            `river:"vsphere_user,attr,optional"`
	VSpherePass             rivertypes.Secret `river:"vsphere_password,attr,optional"`
	ObjectDiscoveryInterval time.Duration     `river:"discovery_interval,attr,optional"`
	EnableExporterMetrics   bool              `river:"enable_exporter_metrics,attr,optional"`
}

// SetToDefault implements river.Defaulter.
func (a *Arguments) SetToDefault() {
	*a = DefaultArguments
}

func (a *Arguments) Convert() *vmware_exporter.Config {
	return &vmware_exporter.Config{
		ChunkSize:               a.ChunkSize,
		CollectConcurrency:      a.CollectConcurrency,
		VSphereURL:              a.VSphereURL,
		VSphereUser:             a.VSphereUser,
		VSpherePass:             config_util.Secret(a.VSpherePass),
		ObjectDiscoveryInterval: a.ObjectDiscoveryInterval,
		EnableExporterMetrics:   a.EnableExporterMetrics,
	}
}
