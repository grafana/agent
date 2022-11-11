package kubernetes_crds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/grafana/agent/component"
	commonConfig "github.com/grafana/agent/component/common/config"
	"github.com/prometheus/prometheus/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promop "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	versioned "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.kubernetes_crds",
		Args:    Config{},
		Exports: struct{}{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args)
		},
	})
}

type Config struct {
	// Local kubeconfig to access cluster
	KubeConfig string `river:"kubeconfig_file,attr,optional"`
	// APIServerConfig allows specifying a host and auth methods to access apiserver.
	// If left empty, Prometheus is assumed to run inside of the cluster
	// and will discover API servers automatically and use the pod's CA certificate
	// and bearer token file at /var/run/secrets/kubernetes.io/serviceaccount/.
	ApiServerConfig *APIServerConfig `river:"api_server,block,optional"`
}

// APIServerConfig defines a host and auth methods to access apiserver.
// More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config
type APIServerConfig struct {
	// Host of apiserver.
	// A valid string consisting of a hostname or IP followed by an optional port number
	Host string `json:"host"`
	// BasicAuth allow an endpoint to authenticate over basic authentication
	BasicAuth *commonConfig.BasicAuth `json:"basicAuth,omitempty"`
	// Bearer token for accessing apiserver.
	BearerToken string `json:"bearerToken,omitempty"`
	// File to read bearer token for accessing apiserver.
	BearerTokenFile string `json:"bearerTokenFile,omitempty"`
	// TLS Config to use for accessing apiserver.
	TLSConfig commonConfig.TLSConfig `json:"tlsConfig,omitempty"`
	// Authorization section for accessing apiserver
	Authorization commonConfig.Authorization `json:"authorization,omitempty"`
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

type Component struct {
	opts      component.Options
	discovery *discovery.Manager

	config           *Config
	cg               configGenerator
	discoveryConfigs map[string]discovery.Configs

	onUpdate chan struct{}
	mut      sync.Mutex
}

func New(o component.Options, args component.Arguments) (*Component, error) {

	c := &Component{
		opts:             o,
		discoveryConfigs: map[string]discovery.Configs{},
		onUpdate:         make(chan struct{}, 1),
	}
	return c, c.Update(args)
}

func (c *Component) clearConfigs(kind string, ns string, name string) {
	prefix := fmt.Sprintf("%s/%s/%s/%d", kind, ns, name)
	for k := range c.discoveryConfigs {
		if strings.HasPrefix(k, prefix) {
			delete(c.discoveryConfigs, k)
		}
	}
}

func (c *Component) OnAddPodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		pmc := c.cg.generatePodMonitorConfig(pm, ep, i)
		c.discoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
	}
	c.discovery.ApplyConfig(c.discoveryConfigs)
}

func (c *Component) OnUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	pm = newObj.(*v1.PodMonitor)
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		pmc := c.cg.generatePodMonitorConfig(pm, ep, i)
		c.discoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
	}
	c.discovery.ApplyConfig(c.discoveryConfigs)
}
func (c *Component) OnDeletePodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	c.discovery.ApplyConfig(c.discoveryConfigs)
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	var cancel func()
	var innerCtx context.Context
	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return nil
		case <-c.onUpdate:
			if cancel != nil {
				cancel()
			}
			innerCtx, cancel = context.WithCancel(ctx)
			c.mut.Lock()
			componentCfg := c.config
			c.mut.Unlock()
			go c.run(innerCtx, componentCfg)
		}
	}
}
func (c *Component) run(ctx context.Context, componentCfg *Config) error {
	// TODO: api server config will usually be default in-cluster config.
	// I am hardcoding this for testing.
	// Presumably, this component will default to in-cluster,
	// and have fields to customize apiserver endpoint.
	// endpoint will be used both for PodMonitor discovery,
	// and the generated Scrape configs to discover pods.
	config, err := componentCfg.restConfig()
	if err != nil {
		return err
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}

	factory := promop.NewSharedInformerFactory(clientset, 5*time.Minute)
	inf := factory.Monitoring().V1().PodMonitors().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.OnAddPodMonitor,
		UpdateFunc: c.OnUpdatePodMonitor,
		DeleteFunc: c.OnDeletePodMonitor,
	})
	factory.Start(ctx.Done())

	c.discovery = discovery.NewManager(ctx, c.opts.Logger, discovery.Name(c.opts.ID))
	go func() {
		err := c.discovery.Run()
		if err != nil {
			// TODO: handle exit better
			log.Fatal(err)
		}
	}()

	go func() {
		for _, up := range <-c.discovery.SyncCh() {
			log.Println(up)
		}
	}()

	// TODO: host filtering hack ?
	<-ctx.Done()
	return nil
}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
	c.mut.Lock()
	cfg := args.(Config)
	c.config = &cfg
	c.cg = configGenerator{
		config: c.config,
	}
	c.discoveryConfigs = map[string]discovery.Configs{}
	c.mut.Unlock()
	select {
	case c.onUpdate <- struct{}{}:
	default:
	}
	return nil
}
