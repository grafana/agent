package kubernetes_crds

import (
	commonConfig "github.com/grafana/agent/component/common/config"
	promConfig "github.com/prometheus/common/config"
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
}

// APIServerConfig defines a host and auth methods to access apiserver.
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
// TODO: river!
type APIServerConfig struct {
	// Host of apiserver.
	// A valid string consisting of a hostname or IP followed by an optional port number
	Host             commonConfig.URL              `river:"host,attr,optional"`
	HTTPClientConfig commonConfig.HTTPClientConfig `river:"http_client_config,block,optional"`
}

func (c *Config) restConfig() (*rest.Config, error) {
	if c.KubeConfig != "" {
		return clientcmd.BuildConfigFromFlags("", c.KubeConfig)
	}
	if c.ApiServerConfig == nil {
		return rest.InClusterConfig()
	}
	rt, err := promConfig.NewRoundTripperFromConfig(*c.ApiServerConfig.HTTPClientConfig.Convert(), "kubernetes_sd")
	if err != nil {
		return nil, err
	}
	return &rest.Config{
		Host:      c.ApiServerConfig.Host.String(),
		Transport: rt,
	}, nil
}
