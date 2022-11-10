package kubernetes_crds

import (
	"context"
	"log"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promop "github.com/prometheus-operator/prometheus-operator/pkg/client/informers/externalversions"
	versioned "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.kubernetes_crds",
		Args:    struct{}{},
		Exports: discovery.Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args)
		},
	})
}

type Component struct {
	opts component.Options
}

func New(o component.Options, args component.Arguments) (*Component, error) {

	c := &Component{
		opts: o,
	}
	return c, c.Update(args)
}

func (c *Component) OnAddPodMonitor(obj interface{}) {
	pm := obj.(*v1.PodMonitor)
	for i, ep := range pm.Spec.PodMetricsEndpoints {
		log.Println(generatePodMonitorConfig(pm, ep, i))
	}
}

func (c *Component) OnUpdatePodMonitor(oldObj, newObj interface{}) {
	log.Println(c)

}
func (c *Component) OnDeletePodMonitor(obj interface{}) {
	log.Println(c)
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/craig/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	factory := promop.NewSharedInformerFactory(clientset, 5*time.Minute)
	inf := factory.Monitoring().V1().PodMonitors().Informer()
	inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.OnAddPodMonitor,
		UpdateFunc: c.OnUpdatePodMonitor,
		DeleteFunc: c.OnDeletePodMonitor,
	})
	factory.Start(ctx.Done())
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

// Update implements component.Compnoent.
func (c *Component) Update(args component.Arguments) error {

	return nil
}
