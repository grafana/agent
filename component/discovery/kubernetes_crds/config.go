package kubernetes_crds

import (
	"log"

	commonConfig "github.com/grafana/agent/component/common/config"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/storage"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	// Local kubeconfig to access cluster
	KubeConfig string `river:"kubeconfig_file,attr,optional"`
	// APIServerConfig allows specifying a host and auth methods to access apiserver.
	// If left empty, Prometheus is assumed to run inside of the cluster
	// and will discover API servers automatically and use the pod's CA certificate
	// and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.
	ApiServerConfig *APIServerConfig `river:"api_server,block,optional"`

	ForwardTo []storage.Appendable `river:"forward_to,attr"`

	// OverrideHonorLabels controls how conflicts in labels are handled
	OverrideHonorLabels     bool
	OverrideHonorTimestamps bool

	EnforcedSampleLimit           *uint64
	EnforcedTargetLimit           *uint64
	EnforcedLabelLimit            *uint64
	EnforcedLabelNameLengthLimit  *uint64
	EnforcedLabelValueLengthLimit *uint64
	//TODO: EnforcedBodySizeLimit         string

	EnforcedNamespaceLabel  string
	ExcludedFromEnforcement []monitoringv1.ObjectReference
}

// APIServerConfig defines a host and auth methods to access apiserver.
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
// TODO: river!
type APIServerConfig struct {
	// Host of apiserver.
	// A valid string consisting of a hostname or IP followed by an optional port number
	// TODO: parse url
	Host commonConfig.URL `json:"host"`
	// BasicAuth allow an endpoint to authenticate over basic authentication
	BasicAuth *commonConfig.BasicAuth `json:"basicAuth,omitempty"`
	// Bearer token for accessing apiserver.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for accessing apiserver.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// TLS Config to use for accessing apiserver.
	TLSConfig *commonConfig.TLSConfig `json:"tlsConfig,omitempty"`
	// Authorization section for accessing apiserver
	Authorization *commonConfig.Authorization `json:"authorization,omitempty"`
}

func (c *Config) restConfig() (*rest.Config, error) {
	if c.KubeConfig != "" {
		return clientcmd.BuildConfigFromFlags("", c.KubeConfig)
	}
	if c.ApiServerConfig == nil {
		return rest.InClusterConfig()
	}
	// TODO
	log.Fatal("Convert apiserverconfig directly")
	return nil, nil
}
