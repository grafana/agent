package operator

import (
	"context"
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/build"
	grafana_v1alpha1 "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/clientutil"
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

func generateMetricsStatefulSetService(cfg *Config, h grafana_v1alpha1.Hierarchy) *v1.Service {
	h = *h.DeepCopy()

	if h.Agent.Spec.PortName == "" {
		h.Agent.Spec.PortName = defaultPortName
	}

	return &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      governingServiceName(h.Agent.Name),
			Namespace: h.Agent.Namespace,
			OwnerReferences: []meta_v1.OwnerReference{{
				APIVersion:         h.Agent.APIVersion,
				Kind:               h.Agent.Kind,
				Name:               h.Agent.Name,
				BlockOwnerDeletion: pointer.Bool(true),
				Controller:         pointer.Bool(true),
				UID:                h.Agent.UID,
			}},
			Labels: cfg.Labels.Merge(map[string]string{
				managedByOperatorLabel: managedByOperatorLabelValue,
				agentNameLabelName:     h.Agent.Name,
				"operated-agent":       "true",
			}),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{{
				Name:       h.Agent.Spec.PortName,
				Port:       8080,
				TargetPort: intstr.FromString(h.Agent.Spec.PortName),
			}},
			Selector: map[string]string{
				"app.kubernetes.io/name": "grafana-agent",
				agentNameLabelName:       h.Agent.Name,
			},
		},
	}
}

func generateMetricsStatefulSet(
	cfg *Config,
	name string,
	h grafana_v1alpha1.Hierarchy,
	shard int32,
) (*apps_v1.StatefulSet, error) {
	h = *h.DeepCopy()

	//
	// Apply defaults to all the fields.
	//

	if h.Agent.Spec.PortName == "" {
		h.Agent.Spec.PortName = defaultPortName
	}

	if h.Agent.Spec.Metrics.Replicas == nil {
		h.Agent.Spec.Metrics.Replicas = &minReplicas
	}

	if h.Agent.Spec.Metrics.Replicas != nil && *h.Agent.Spec.Metrics.Replicas < 0 {
		intZero := int32(0)
		h.Agent.Spec.Metrics.Replicas = &intZero
	}
	if h.Agent.Spec.Resources.Requests == nil {
		h.Agent.Spec.Resources.Requests = v1.ResourceList{}
	}

	spec, err := generateMetricsStatefulSetSpec(cfg, name, h, shard)
	if err != nil {
		return nil, err
	}

	// Don't transfer any kubectl annotations to the statefulset so it doesn't
	// get pruned by kubectl.
	annotations := make(map[string]string)
	for k, v := range h.Agent.Annotations {
		if !strings.HasPrefix(k, "kubectl.kubernetes.io/") {
			annotations[k] = v
		}
	}

	labels := make(map[string]string)
	for k, v := range spec.Template.Labels {
		labels[k] = v
	}
	labels[agentNameLabelName] = h.Agent.Name
	labels[agentTypeLabel] = "metrics"
	labels[managedByOperatorLabel] = managedByOperatorLabelValue

	ss := &apps_v1.StatefulSet{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        name,
			Namespace:   h.Agent.Namespace,
			Labels:      labels,
			Annotations: annotations,
			OwnerReferences: []meta_v1.OwnerReference{{
				APIVersion:         h.Agent.APIVersion,
				Kind:               h.Agent.Kind,
				BlockOwnerDeletion: pointer.Bool(true),
				Controller:         pointer.Bool(true),
				Name:               h.Agent.Name,
				UID:                h.Agent.UID,
			}},
		},
		Spec: *spec,
	}

	// TODO(rfratto): Prometheus Operator has an input hash annotation added here,
	// which combines the hash of the statefulset, config to the operator, rule
	// config map names (unused here), and the previous statefulset (if any).
	//
	// This is used to skip re-applying an unchanged statefulset. Do we need this?

	if len(h.Agent.Spec.ImagePullSecrets) > 0 {
		ss.Spec.Template.Spec.ImagePullSecrets = h.Agent.Spec.ImagePullSecrets
	}

	storageSpec := h.Agent.Spec.Storage
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

	ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, h.Agent.Spec.Volumes...)

	return ss, nil
}

func generateMetricsStatefulSetSpec(
	cfg *Config,
	name string,
	h grafana_v1alpha1.Hierarchy,
	shard int32,
) (*apps_v1.StatefulSetSpec, error) {

	shards := minShards
	if reqShards := h.Agent.Spec.Metrics.Shards; reqShards != nil && *reqShards > 1 {
		shards = *reqShards
	}

	useVersion := h.Agent.Spec.Version
	if useVersion == "" {
		useVersion = DefaultAgentVersion
	}
	imagePath := fmt.Sprintf("%s:%s", DefaultAgentBaseImage, useVersion)
	if h.Agent.Spec.Image != nil && *h.Agent.Spec.Image != "" {
		imagePath = *h.Agent.Spec.Image
	}

	agentArgs := []string{
		"-config.file=/var/lib/grafana-agent/config/agent.yml",
		"-config.expand-env=true",
		"-reload-port=8081",
	}

	// NOTE(rfratto): the Prometheus Operator supports a ListenLocal to prevent a
	// service from being created. Given the intent is that Agents can connect to
	// each other, ListenLocal isn't currently supported and we always create a port.
	ports := []v1.ContainerPort{{
		Name:          h.Agent.Spec.PortName,
		ContainerPort: 8080,
		Protocol:      v1.ProtocolTCP,
	}}

	volumes := []v1.Volume{
		{
			Name: "config",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: fmt.Sprintf("%s-config", h.Agent.Name),
				},
			},
		},
		{
			// We need a separate volume for storing the rendered config with
			// environment variables replaced. While the Agent supports environment
			// variable substitution, the value for __replica__ can only be
			// determined at runtime. We use a dedicated container for both config
			// reloading and rendering.
			Name: "config-out",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "secrets",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: fmt.Sprintf("%s-secrets", h.Agent.Name),
				},
			},
		},
	}

	walVolumeName := fmt.Sprintf("%s-wal", name)
	if h.Agent.Spec.Storage != nil {
		if h.Agent.Spec.Storage.VolumeClaimTemplate.Name != "" {
			walVolumeName = h.Agent.Spec.Storage.VolumeClaimTemplate.Name
		}
	}

	volumeMounts := []v1.VolumeMount{
		{
			Name:      "config",
			ReadOnly:  true,
			MountPath: "/var/lib/grafana-agent/config-in",
		},
		{
			Name:      "config-out",
			MountPath: "/var/lib/grafana-agent/config",
		},
		{
			Name:      walVolumeName,
			ReadOnly:  false,
			MountPath: "/var/lib/grafana-agent/data",
		},
		{
			Name:      "secrets",
			ReadOnly:  true,
			MountPath: "/var/lib/grafana-agent/secrets",
		},
	}
	volumeMounts = append(volumeMounts, h.Agent.Spec.VolumeMounts...)

	for _, s := range h.Agent.Spec.Secrets {
		volumes = append(volumes, v1.Volume{
			Name: clientutil.SanitizeVolumeName("secret-" + s),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{SecretName: s},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      clientutil.SanitizeVolumeName("secret-" + s),
			ReadOnly:  true,
			MountPath: "/var/lib/grafana-agent/secrets",
		})
	}

	for _, c := range h.Agent.Spec.ConfigMaps {
		volumes = append(volumes, v1.Volume{
			Name: clientutil.SanitizeVolumeName("configmap-" + c),
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{Name: c},
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      clientutil.SanitizeVolumeName("configmap-" + c),
			ReadOnly:  true,
			MountPath: "/var/lib/grafana-agent/configmaps",
		})
	}

	podAnnotations := map[string]string{}
	podLabels := map[string]string{}
	podSelectorLabels := map[string]string{
		"app.kubernetes.io/name":     "grafana-agent",
		"app.kubernetes.io/version":  build.Version,
		"app.kubernetes.io/instance": h.Agent.Name,
		"grafana-agent":              h.Agent.Name,
		managedByOperatorLabel:       managedByOperatorLabelValue,
		shardLabelName:               fmt.Sprintf("%d", shard),
		agentNameLabelName:           h.Agent.Name,
		agentTypeLabel:               "metrics",
	}
	if h.Agent.Spec.PodMetadata != nil {
		for k, v := range h.Agent.Spec.PodMetadata.Labels {
			podLabels[k] = v
		}
		for k, v := range h.Agent.Spec.PodMetadata.Annotations {
			podAnnotations[k] = v
		}
	}
	for k, v := range podSelectorLabels {
		podLabels[k] = v
	}

	podAnnotations["kubectl.kubernetes.io/default-container"] = "grafana-agent"

	var (
		finalSelectorLabels = cfg.Labels.Merge(podSelectorLabels)
		finalLabels         = cfg.Labels.Merge(podLabels)
	)

	envVars := []v1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
			},
		},
		{
			Name:  "SHARD",
			Value: fmt.Sprintf("%d", shard),
		},
		{
			Name:  "SHARDS",
			Value: fmt.Sprintf("%d", shards),
		},
	}

	operatorContainers := []v1.Container{
		{
			Name:         "config-reloader",
			Image:        "quay.io/prometheus-operator/prometheus-config-reloader:v0.47.0",
			VolumeMounts: volumeMounts,
			Env:          envVars,
			Args: []string{
				"--config-file=/var/lib/grafana-agent/config-in/agent.yml",
				"--config-envsubst-file=/var/lib/grafana-agent/config/agent.yml",

				"--watch-interval=1m",
				"--statefulset-ordinal-from-envvar=POD_NAME",

				// Use specifically the reload-port for reloading, since the primary
				// server can shut down in between reloads.
				"--reload-url=http://127.0.0.1:8081/-/reload",
			},
		},
		{
			Name:         "grafana-agent",
			Image:        imagePath,
			Ports:        ports,
			Args:         agentArgs,
			VolumeMounts: volumeMounts,
			Env:          envVars,
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/-/ready",
						Port: intstr.FromString(h.Agent.Spec.PortName),
					},
				},
				TimeoutSeconds:   probeTimeoutSeconds,
				PeriodSeconds:    5,
				FailureThreshold: 120, // Allow up to 10m on startup for data recovery
			},
			Resources:                h.Agent.Spec.Resources,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
	}

	containers, err := clientutil.MergePatchContainers(operatorContainers, h.Agent.Spec.Containers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge containers spec: %w", err)
	}

	return &apps_v1.StatefulSetSpec{
		ServiceName:         governingServiceName(h.Agent.Name),
		Replicas:            h.Agent.Spec.Metrics.Replicas,
		PodManagementPolicy: apps_v1.ParallelPodManagement,
		UpdateStrategy: apps_v1.StatefulSetUpdateStrategy{
			Type: apps_v1.RollingUpdateStatefulSetStrategyType,
		},
		Selector: &meta_v1.LabelSelector{
			MatchLabels: finalSelectorLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: meta_v1.ObjectMeta{
				Labels:      finalLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				Containers:                    containers,
				InitContainers:                h.Agent.Spec.InitContainers,
				SecurityContext:               h.Agent.Spec.SecurityContext,
				ServiceAccountName:            h.Agent.Spec.ServiceAccountName,
				NodeSelector:                  h.Agent.Spec.NodeSelector,
				PriorityClassName:             h.Agent.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: pointer.Int64(4800),
				Volumes:                       volumes,
				Tolerations:                   h.Agent.Spec.Tolerations,
				Affinity:                      h.Agent.Spec.Affinity,
				TopologySpreadConstraints:     h.Agent.Spec.TopologySpreadConstraints,
			},
		},
	}, nil
}
