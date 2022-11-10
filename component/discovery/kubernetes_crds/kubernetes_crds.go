package kubernetes_crds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/grafana/agent/component"
	"github.com/prometheus/prometheus/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promop "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	versioned "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.kubernetes_crds",
		Args:    struct{}{},
		Exports: struct{}{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args)
		},
	})
}

type Component struct {
	opts             component.Options
	discovery        *discovery.Manager
	discoveryConfigs map[string]discovery.Configs
}

func New(o component.Options, args component.Arguments) (*Component, error) {

	c := &Component{
		opts:             o,
		discoveryConfigs: map[string]discovery.Configs{},
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
		pmc := generatePodMonitorConfig(pm, ep, i)
		c.discoveryConfigs[pmc.JobName] = pmc.ServiceDiscoveryConfigs
	}
	c.discovery.ApplyConfig(c.discoveryConfigs)
}

func (c *Component) OnUpdatePodMonitor(oldObj, newObj interface{}) {
	pm := oldObj.(*v1.PodMonitor)
	c.clearConfigs("podMonitor", pm.Namespace, pm.Name)
	pm = newObj.(*v1.PodMonitor)
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		pmc := generatePodMonitorConfig(pm, ep, i)
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
	// TODO: api server config will usually be default in-cluster config.
	// I am hardcoding this for testing.
	// Presumably, this component will default to in-cluster,
	// and have fields to customize apiserver endpoint.
	// endpoint will be used both for PodMonitor discovery,
	// and the generated Scrape configs to discover pods.
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/craig/.kube/config")
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

	// TODO: host filtering hack ?
	<-ctx.Done()
	return nil
}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {
	return nil
}
