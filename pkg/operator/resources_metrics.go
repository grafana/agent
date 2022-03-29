package operator

import (
	"context"
	"fmt"
	"strings"

	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	prom_operator "github.com/prometheus-operator/prometheus-operator/pkg/operator"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultPortName = "http-metrics"
)

var (
	minShards                   int32 = 1
	minReplicas                 int32 = 1
	managedByOperatorLabel            = "app.kubernetes.io/managed-by"
	managedByOperatorLabelValue       = "grafana-agent-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
	shardLabelName            = "operator.agent.grafana.com/shard"
	agentNameLabelName        = "operator.agent.grafana.com/name"
	agentTypeLabel            = "operator.agent.grafana.com/type"
	probeTimeoutSeconds int32 = 3
)

// deleteManagedResource deletes a managed resource. Ignores resources that are
// not managed.
func deleteManagedResource(ctx context.Context, cli client.Client, key client.ObjectKey, o client.Object) error {
	err := cli.Get(ctx, key, o)
	if k8s_errors.IsNotFound(err) || !isManagedResource(o) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to find stale resource %s: %w", key, err)
	}
	err = cli.Delete(ctx, o)
	if err != nil {
		return fmt.Errorf("failed to delete stale resource %s: %w", key, err)
	}
	return nil
}

// isManagedResource returns true if the given object has a managed-by
// grafana-agent-operator label.
func isManagedResource(obj client.Object) bool {
	labelValue := obj.GetLabels()[managedByOperatorLabel]
	return labelValue == managedByOperatorLabelValue
}

func governingServiceName(agentName string) string {
	return fmt.Sprintf("%s-operated", agentName)
}

func generateMetricsStatefulSetService(cfg *Config, d gragent.Deployment) *v1.Service {
	d = *d.DeepCopy()

	if d.Agent.Spec.PortName == "" {
		d.Agent.Spec.PortName = defaultPortName
	}

	return &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      governingServiceName(d.Agent.Name),
			Namespace: d.Agent.Namespace,
			OwnerReferences: []meta_v1.OwnerReference{{
				APIVersion:         d.Agent.APIVersion,
				Kind:               d.Agent.Kind,
				Name:               d.Agent.Name,
				BlockOwnerDeletion: pointer.Bool(true),
				Controller:         pointer.Bool(true),
				UID:                d.Agent.UID,
			}},
			Labels: cfg.Labels.Merge(map[string]string{
				managedByOperatorLabel: managedByOperatorLabelValue,
				agentNameLabelName:     d.Agent.Name,
				"operated-agent":       "true",
			}),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{{
				Name:       d.Agent.Spec.PortName,
				Port:       8080,
				TargetPort: intstr.FromString(d.Agent.Spec.PortName),
			}},
			Selector: map[string]string{
				"app.kubernetes.io/name": "grafana-agent",
				agentNameLabelName:       d.Agent.Name,
			},
		},
	}
}

func generateMetricsStatefulSet(
	cfg *Config,
	name string,
	d gragent.Deployment,
	shard int32,
) (*apps_v1.StatefulSet, error) {

	d = *d.DeepCopy()

	//
	// Apply defaults to all the fields.
	//

	if d.Agent.Spec.PortName == "" {
		d.Agent.Spec.PortName = defaultPortName
	}

	if d.Agent.Spec.Metrics.Replicas == nil {
		d.Agent.Spec.Metrics.Replicas = &minReplicas
	}

	if d.Agent.Spec.Metrics.Replicas != nil && *d.Agent.Spec.Metrics.Replicas < 0 {
		intZero := int32(0)
		d.Agent.Spec.Metrics.Replicas = &intZero
	}
	if d.Agent.Spec.Resources.Requests == nil {
		d.Agent.Spec.Resources.Requests = v1.ResourceList{}
	}

	spec, err := generateMetricsStatefulSetSpec(cfg, name, d, shard)
	if err != nil {
		return nil, err
	}

	// Don't transfer any kubectl annotations to the statefulset so it doesn't
	// get pruned by kubectl.
	annotations := make(map[string]string)
	for k, v := range d.Agent.Annotations {
		if !strings.HasPrefix(k, "kubectl.kubernetes.io/") {
			annotations[k] = v
		}
	}

	labels := make(map[string]string)
	for k, v := range spec.Template.Labels {
		labels[k] = v
	}
	labels[agentNameLabelName] = d.Agent.Name
	labels[agentTypeLabel] = "metrics"
	labels[managedByOperatorLabel] = managedByOperatorLabelValue

	ss := &apps_v1.StatefulSet{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        name,
			Namespace:   d.Agent.Namespace,
			Labels:      labels,
			Annotations: annotations,
			OwnerReferences: []meta_v1.OwnerReference{{
				APIVersion:         d.Agent.APIVersion,
				Kind:               d.Agent.Kind,
				BlockOwnerDeletion: pointer.Bool(true),
				Controller:         pointer.Bool(true),
				Name:               d.Agent.Name,
				UID:                d.Agent.UID,
			}},
		},
		Spec: *spec,
	}

	// TODO(rfratto): Prometheus Operator has an input hash annotation added here,
	// which combines the hash of the statefulset, config to the operator, rule
	// config map names (unused here), and the previous statefulset (if any).
	//
	// This is used to skip re-applying an unchanged statefulset. Do we need this?

	if len(d.Agent.Spec.ImagePullSecrets) > 0 {
		ss.Spec.Template.Spec.ImagePullSecrets = d.Agent.Spec.ImagePullSecrets
	}

	storageSpec := d.Agent.Spec.Storage
	if storageSpec == nil {
		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, v1.Volume{
			Name: fmt.Sprintf("%s-wal", name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, v1.Volume{
			Name: fmt.Sprintf("%s-wal", name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else {
		pvcTemplate := prom_operator.MakeVolumeClaimTemplate(storageSpec.VolumeClaimTemplate)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = fmt.Sprintf("%s-wal", name)
		}
		if storageSpec.VolumeClaimTemplate.Spec.AccessModes == nil {
			pvcTemplate.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		} else {
			pvcTemplate.Spec.AccessModes = storageSpec.VolumeClaimTemplate.Spec.AccessModes
		}
		pvcTemplate.Spec.Resources = storageSpec.VolumeClaimTemplate.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.VolumeClaimTemplate.Spec.Selector
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, *pvcTemplate)
	}

	ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, d.Agent.Spec.Volumes...)

	return ss, nil
}

func generateMetricsStatefulSetSpec(
	cfg *Config,
	name string,
	d gragent.Deployment,
	shard int32,
) (*apps_v1.StatefulSetSpec, error) {

	shards := minShards
	if reqShards := d.Agent.Spec.Metrics.Shards; reqShards != nil && *reqShards > 1 {
		shards = *reqShards
	}

	walVolumeName := fmt.Sprintf("%s-wal", name)
	if d.Agent.Spec.Storage != nil {
		if d.Agent.Spec.Storage.VolumeClaimTemplate.Name != "" {
			walVolumeName = d.Agent.Spec.Storage.VolumeClaimTemplate.Name
		}
	}

	opts := podTemplateOptions{
		ExtraSelectorLabels: map[string]string{
			shardLabelName: fmt.Sprintf("%d", shard),
			agentTypeLabel: "metrics",
		},
		ExtraVolumeMounts: []v1.VolumeMount{{
			Name:      walVolumeName,
			ReadOnly:  false,
			MountPath: "/var/lib/grafana-agent/data",
		}},
		ExtraEnvVars: []v1.EnvVar{
			{
				Name:  "SHARD",
				Value: fmt.Sprintf("%d", shard),
			},
			{
				Name:  "SHARDS",
				Value: fmt.Sprintf("%d", shards),
			},
		},
	}

	templateSpec, selector, err := generatePodTemplate(cfg, name, d, opts)
	if err != nil {
		return nil, err
	}

	return &apps_v1.StatefulSetSpec{
		ServiceName:         governingServiceName(d.Agent.Name),
		Replicas:            d.Agent.Spec.Metrics.Replicas,
		PodManagementPolicy: apps_v1.ParallelPodManagement,
		UpdateStrategy: apps_v1.StatefulSetUpdateStrategy{
			Type: apps_v1.RollingUpdateStatefulSetStrategyType,
		},
		Selector: selector,
		Template: templateSpec,
	}, nil
}
