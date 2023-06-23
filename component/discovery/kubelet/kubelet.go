// Package kubelet implements a discovery.kubelet component.
package kubelet

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/discovery"
	discoveryk8s "github.com/grafana/agent/component/discovery/kubernetes"
	commonConfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/refresh"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/strutil"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultKubeletRefreshInterval = time.Minute

	metaLabelPrefix               = model.MetaLabelPrefix + "kubernetes_"
	namespaceLabel                = metaLabelPrefix + "namespace"
	nodeNameLabel                 = metaLabelPrefix + "node_name"
	presentValue                  = model.LabelValue("true")
	podNameLabel                  = metaLabelPrefix + "pod_name"
	podIPLabel                    = metaLabelPrefix + "pod_ip"
	podContainerNameLabel         = metaLabelPrefix + "pod_container_name"
	podContainerImageLabel        = metaLabelPrefix + "pod_container_image"
	podContainerPortNameLabel     = metaLabelPrefix + "pod_container_port_name"
	podContainerPortNumberLabel   = metaLabelPrefix + "pod_container_port_number"
	podContainerPortProtocolLabel = metaLabelPrefix + "pod_container_port_protocol"
	podContainerIsInit            = metaLabelPrefix + "pod_container_init"
	podReadyLabel                 = metaLabelPrefix + "pod_ready"
	podPhaseLabel                 = metaLabelPrefix + "pod_phase"
	podLabelPrefix                = metaLabelPrefix + "pod_label_"
	podLabelPresentPrefix         = metaLabelPrefix + "pod_labelpresent_"
	podAnnotationPrefix           = metaLabelPrefix + "pod_annotation_"
	podAnnotationPresentPrefix    = metaLabelPrefix + "pod_annotationpresent_"
	podNodeNameLabel              = metaLabelPrefix + "pod_node_name"
	podHostIPLabel                = metaLabelPrefix + "pod_host_ip"
	podUID                        = metaLabelPrefix + "pod_uid"
	podControllerKind             = metaLabelPrefix + "pod_controller_kind"
	podControllerName             = metaLabelPrefix + "pod_controller_name"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.kubelet",
		Args:    Arguments{},
		Exports: discovery.Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments configures the discovery.kubernetes component.
type Arguments struct {
	KubeletURL       config.URL                        `river:"kubelet_url,attr"`
	Interval         time.Duration                     `river:"refresh_interval,attr,optional"`
	HTTPClientConfig config.HTTPClientConfig           `river:",squash"`
	Namespaces       []string                          `river:"namespaces,block"`
	AttachMetadata   discoveryk8s.AttachMetadataConfig `river:"attach_metadata,block,optional"`
}

// New returns a new instance of a discovery.kubelet component.
func New(opts component.Options, args Arguments) (*discovery.Component, error) {
	return discovery.New(opts, args, func(args component.Arguments) (discovery.Discoverer, error) {
		newArgs := args.(Arguments)
		kubeletDiscovery, err := NewKubeletDiscovery(newArgs)
		if err != nil {
			return nil, err
		}
		interval := defaultKubeletRefreshInterval
		if newArgs.Interval != 0 {
			interval = newArgs.Interval
		}
		return refresh.NewDiscovery(opts.Logger, "kubelet", interval, kubeletDiscovery.Refresh), nil
	})
}

type Discovery struct {
	client           *http.Client
	url              string
	addNodeMetadata  bool
	targetNamespaces []string
}

func NewKubeletDiscovery(args Arguments) (*Discovery, error) {
	transport, err := commonConfig.NewRoundTripperFromConfig(*args.HTTPClientConfig.Convert(), "kubelet_sd")
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	// ensure the path is the kubelet pods endpoint
	args.KubeletURL.Path = "/pods"
	return &Discovery{
		client:           client,
		url:              args.KubeletURL.String(),
		addNodeMetadata:  args.AttachMetadata.Node,
		targetNamespaces: args.Namespaces,
	}, nil
}

func (d *Discovery) Refresh(_ context.Context) ([]*targetgroup.Group, error) {
	// Create a new GET request to the kubelet API server
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating kubelet GET pods request: %v", err)
	}

	// Send the request to the kubelet
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending kublet GET pods request: %v", err)
	}

	// Read the response body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Unmarshal the response body into a list of pods
	var pods []*v1.Pod
	if err := json.Unmarshal(body, &pods); err != nil {
		return nil, fmt.Errorf("error unmarshaling response body: %v", err)
	}

	// Create a list of targets from the pods
	var targetGroups []*targetgroup.Group
	for _, pod := range pods {
		// Skip pods that are not in the one of the desired namespaces
		if len(d.targetNamespaces) > 0 && !d.podInTargetNamespaces(pod) {
			continue
		}
		targetGroups = append(targetGroups, d.buildPodTargetGroup(pod))
	}

	return targetGroups, nil
}

func (d *Discovery) buildPodTargetGroup(pod *v1.Pod) *targetgroup.Group {
	tg := &targetgroup.Group{
		Source: podSource(pod),
	}
	// PodIP can be empty when a pod is starting or has been evicted.
	if len(pod.Status.PodIP) == 0 {
		return tg
	}

	tg.Labels = podLabels(pod)
	tg.Labels[namespaceLabel] = lv(pod.Namespace)
	if d.addNodeMetadata {
		tg.Labels[nodeNameLabel] = lv(pod.Spec.NodeName)
	}

	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	for i, c := range containers {
		isInit := i >= len(pod.Spec.Containers)
		for _, port := range c.Ports {
			ports := strconv.FormatUint(uint64(port.ContainerPort), 10)
			addr := net.JoinHostPort(pod.Status.PodIP, ports)

			tg.Targets = append(tg.Targets, model.LabelSet{
				model.AddressLabel:            lv(addr),
				podContainerNameLabel:         lv(c.Name),
				podContainerImageLabel:        lv(c.Image),
				podContainerPortNumberLabel:   lv(ports),
				podContainerPortNameLabel:     lv(port.Name),
				podContainerPortProtocolLabel: lv(string(port.Protocol)),
				podContainerIsInit:            lv(strconv.FormatBool(isInit)),
			})
		}
	}

	return tg
}

func (d *Discovery) podInTargetNamespaces(pod *v1.Pod) bool {
	for _, ns := range d.targetNamespaces {
		if pod.Namespace == ns {
			return true
		}
	}
	return false
}

func podSource(pod *v1.Pod) string {
	return podSourceFromNamespaceAndName(pod.Namespace, pod.Name)
}

func podSourceFromNamespaceAndName(namespace, name string) string {
	return "pod/" + namespace + "/" + name
}

func podLabels(pod *v1.Pod) model.LabelSet {
	ls := model.LabelSet{
		podNameLabel:     lv(pod.ObjectMeta.Name),
		podIPLabel:       lv(pod.Status.PodIP),
		podReadyLabel:    podReady(pod),
		podPhaseLabel:    lv(string(pod.Status.Phase)),
		podNodeNameLabel: lv(pod.Spec.NodeName),
		podHostIPLabel:   lv(pod.Status.HostIP),
		podUID:           lv(string(pod.ObjectMeta.UID)),
	}

	createdBy := metav1.GetControllerOf(pod)
	if createdBy != nil {
		if createdBy.Kind != "" {
			ls[podControllerKind] = lv(createdBy.Kind)
		}
		if createdBy.Name != "" {
			ls[podControllerName] = lv(createdBy.Name)
		}
	}

	for k, v := range pod.Labels {
		ln := strutil.SanitizeLabelName(k)
		ls[model.LabelName(podLabelPrefix+ln)] = lv(v)
		ls[model.LabelName(podLabelPresentPrefix+ln)] = presentValue
	}

	for k, v := range pod.Annotations {
		ln := strutil.SanitizeLabelName(k)
		ls[model.LabelName(podAnnotationPrefix+ln)] = lv(v)
		ls[model.LabelName(podAnnotationPresentPrefix+ln)] = presentValue
	}

	return ls
}

func lv(s string) model.LabelValue {
	return model.LabelValue(s)
}

func podReady(pod *v1.Pod) model.LabelValue {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == v1.PodReady {
			return lv(strings.ToLower(string(cond.Status)))
		}
	}
	return lv(strings.ToLower(string(v1.ConditionUnknown)))
}
