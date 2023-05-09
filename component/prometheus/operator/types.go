package operator

import (
	"time"

	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/common/kubernetes"
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
}

// Clustering holds values that configure clustering-specific behavior.
type Clustering struct {
	// TODO(@tpaschalis) Move this block to a shared place for all components using clustering.
	Enabled bool `river:"enabled,attr"`
}

var DefaultArguments = Arguments{
	Client: kubernetes.ClientArguments{
		HTTPClientConfig: config.DefaultHTTPClientConfig,
	},
}

func (args *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*args = DefaultArguments

	type arguments Arguments
	if err := f((*arguments)(args)); err != nil {
		return err
	}

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
	Namespace      string    `river:"namespace,attr"`
	Name           string    `river:"name,attr"`
	LastReconcile  time.Time `river:"last_reconcile,attr,optional"`
	ReconcileError string    `river:"reconcile_error,attr,optional"`
}
