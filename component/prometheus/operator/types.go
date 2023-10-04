package operator

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/common/kubernetes"
	flow_relabel "github.com/grafana/agent/component/common/relabel"
	"github.com/grafana/agent/component/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	apiv1 "k8s.io/api/core/v1"
)

type Arguments struct {

	// Client settings to connect to Kubernetes.
	Client kubernetes.ClientArguments `river:"client,block,optional"`

	ForwardTo []storage.Appendable `river:"forward_to,attr"`

	// Namespaces to search for monitor resources. Empty implies All namespaces
	Namespaces []string `river:"namespaces,attr,optional"`

	// LabelSelector allows filtering discovered monitor resources by labels
	LabelSelector *config.LabelSelector `river:"selector,block,optional"`

	Clustering Clustering `river:"clustering,block,optional"`

	RelabelConfigs []*flow_relabel.Config `river:"rule,block,optional"`

	Scrape ScrapeOptions `river:"scrape,block,optional"`
}

// Clustering holds values that configure clustering-specific behavior.
type Clustering struct {
	// TODO(@tpaschalis) Move this block to a shared place for all components using clustering.
	Enabled bool `river:"enabled,attr"`
}

// ScrapeOptions holds values that configure scraping behavior.
type ScrapeOptions struct {
	// DefaultScrapeInterval is the default interval to scrape targets.
	DefaultScrapeInterval time.Duration `river:"default_scrape_interval,attr,optional"`

	// DefaultScrapeTimeout is the default timeout to scrape targets.
	DefaultScrapeTimeout time.Duration `river:"default_scrape_timeout,attr,optional"`
}

var DefaultArguments = Arguments{
	Client: kubernetes.ClientArguments{
		HTTPClientConfig: config.DefaultHTTPClientConfig,
	},
}

// SetToDefault implements river.Defaulter.
func (args *Arguments) SetToDefault() {
	*args = DefaultArguments
}

// Validate implements river.Validator.
func (args *Arguments) Validate() error {
	if len(args.Namespaces) == 0 {
		args.Namespaces = []string{apiv1.NamespaceAll}
	}
	return nil
}

type DebugInfo struct {
	DiscoveredCRDs []*DiscoveredResource `river:"crds,block"`
	Targets        []scrape.TargetStatus `river:"targets,block,optional"`
}

type DiscoveredResource struct {
	Namespace        string    `river:"namespace,attr"`
	Name             string    `river:"name,attr"`
	LastReconcile    time.Time `river:"last_reconcile,attr,optional"`
	ReconcileError   string    `river:"reconcile_error,attr,optional"`
	ScrapeConfigsURL string    `river:"scrape_configs_url,attr,optional"`
}
