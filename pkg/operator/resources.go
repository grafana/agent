package operator

import (
	"fmt"
	"strings"

	"github.com/grafana/agent/pkg/build"
	"github.com/grafana/agent/pkg/operator/clientutil"
	"github.com/grafana/agent/pkg/operator/config"
	prom_operator "github.com/prometheus-operator/prometheus-operator/pkg/operator"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	governingServiceName = "grafana-agent-operated"
	defaultPortName      = "http-metrics"
)

var (
	minShards                   int32 = 1
	minReplicas                 int32 = 1
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "grafana-agent-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}
	shardLabelName            = "operator.agent.grafana.com/shard"
	agentNameLabelName        = "operator.agent.grafana.com/name"
	probeTimeoutSeconds int32 = 3
)

func generateStatefulSetService(cfg *Config, d config.Deployment) *v1.Service {
	d = *d.DeepCopy()

	if d.Agent.Spec.PortName == "" {
		d.Agent.Spec.PortName = defaultPortName
	}

	boolTrue := true

	return &v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      governingServiceName,
			Namespace: d.Agent.ObjectMeta.Namespace,
			OwnerReferences: []meta_v1.OwnerReference{{
				APIVersion:         d.Agent.APIVersion,
				Kind:               d.Agent.Kind,
				Name:               d.Agent.Name,
				BlockOwnerDeletion: &boolTrue,
				Controller:         &boolTrue,
				UID:                d.Agent.UID,
			}},
			Labels: cfg.Labels.Merge(map[string]string{
				"operated-agent": "true",
			}),
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{{
				Name:       d.Agent.Spec.PortName,
				Port:       9090,
				TargetPort: intstr.FromString(d.Agent.Spec.PortName),
			}},
			Selector: map[string]string{
				"app.kubernetes.io/name": "grafana-agent",
			},
		},
	}
}

func generateStatefulSet(
	cfg *Config,
	name string,
	d config.Deployment,
	shard int32,
) (*apps_v1.StatefulSet, error) {
	d = *d.DeepCopy()

	//
	// Apply defaults to all the fields.
	//

	if d.Agent.Spec.PortName == "" {
		d.Agent.Spec.PortName = defaultPortName
	}

	if d.Agent.Spec.Prometheus.Replicas == nil {
		d.Agent.Spec.Prometheus.Replicas = &minReplicas
	}

	if d.Agent.Spec.Prometheus.Replicas != nil && *d.Agent.Spec.Prometheus.Replicas < 0 {
		intZero := int32(0)
		d.Agent.Spec.Prometheus.Replicas = &intZero
	}
	if d.Agent.Spec.Resources.Requests == nil {
		d.Agent.Spec.Resources.Requests = v1.ResourceList{}
	}

	spec, err := generateStatefulSetSpec(cfg, name, d, shard)
	if err != nil {
		return nil, err
	}

	// Don't transfer any kubectl annotations to the statefulset so it doesn't
	// get pruned by kubectl.
	annotations := make(map[string]string)
	for k, v := range d.Agent.ObjectMeta.Annotations {
		if !strings.HasPrefix(k, "kubectl.kubernetes.io/") {
			annotations[k] = v
		}
	}

	labels := make(map[string]string)
	for k, v := range spec.Template.Labels {
		labels[k] = v
	}
	labels[agentNameLabelName] = d.Agent.Name

	boolTrue := true

	ss := &apps_v1.StatefulSet{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        name,
			Namespace:   d.Agent.Namespace,
			Labels:      labels,
			Annotations: annotations,
			OwnerReferences: []meta_v1.OwnerReference{{
				APIVersion:         d.Agent.APIVersion,
				Kind:               d.Agent.Kind,
				BlockOwnerDeletion: &boolTrue,
				Controller:         &boolTrue,
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

func generateStatefulSetSpec(
	cfg *Config,
	name string,
	d config.Deployment,
	shard int32,
) (*apps_v1.StatefulSetSpec, error) {

	shards := minShards
	if reqShards := d.Agent.Spec.Prometheus.Shards; reqShards != nil && *reqShards > 1 {
		shards = *reqShards
	}

	terminationGracePeriodSeconds := int64(4800)

	imagePath := fmt.Sprintf("%s:%s", DefaultAgentBaseImage, d.Agent.Spec.Version)
	if d.Agent.Spec.Image != nil && *d.Agent.Spec.Image != "" {
		imagePath = *d.Agent.Spec.Image
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
		Name:          d.Agent.Spec.PortName,
		ContainerPort: 8080,
		Protocol:      v1.ProtocolTCP,
	}}

	volumes := []v1.Volume{
		{
			Name: "config",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: fmt.Sprintf("%s-config", d.Agent.Name),
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
					SecretName: fmt.Sprintf("%s-secrets", d.Agent.Name),
				},
			},
		},
	}

	walVolumeName := fmt.Sprintf("%s-wal", name)
	if d.Agent.Spec.Storage != nil {
		if d.Agent.Spec.Storage.VolumeClaimTemplate.Name != "" {
			walVolumeName = d.Agent.Spec.Storage.VolumeClaimTemplate.Name
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
	volumeMounts = append(volumeMounts, d.Agent.Spec.VolumeMounts...)

	for _, s := range d.Agent.Spec.Secrets {
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

	for _, c := range d.Agent.Spec.ConfigMaps {
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
		"app.kubernetes.io/name":       "grafana-agent",
		"app.kubernetes.io/version":    build.Version,
		"app.kubernetes.io/managed-by": "grafana-agent-operator",
		"app.kubernetes.io/instance":   d.Agent.Name,
		"grafana-agent":                d.Agent.Name,
		shardLabelName:                 fmt.Sprintf("%d", shard),
		agentNameLabelName:             d.Agent.Name,
	}
	if d.Agent.Spec.PodMetadata != nil {
		for k, v := range d.Agent.Spec.PodMetadata.Labels {
			podLabels[k] = v
		}
		for k, v := range d.Agent.Spec.PodMetadata.Annotations {
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
						Port: intstr.FromString(d.Agent.Spec.PortName),
					},
				},
				TimeoutSeconds:   probeTimeoutSeconds,
				PeriodSeconds:    5,
				FailureThreshold: 120, // Allow up to 10m on startup for data recovery
			},
			Resources:                d.Agent.Spec.Resources,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
	}

	containers, err := clientutil.MergePatchContainers(operatorContainers, d.Agent.Spec.Containers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge containers spec: %w", err)
	}

	return &apps_v1.StatefulSetSpec{
		ServiceName:         governingServiceName,
		Replicas:            d.Agent.Spec.Prometheus.Replicas,
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
				InitContainers:                d.Agent.Spec.InitContainers,
				SecurityContext:               d.Agent.Spec.SecurityContext,
				ServiceAccountName:            d.Agent.Spec.ServiceAccountName,
				NodeSelector:                  d.Agent.Spec.NodeSelector,
				PriorityClassName:             d.Agent.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
				Volumes:                       volumes,
				Tolerations:                   d.Agent.Spec.Tolerations,
				Affinity:                      d.Agent.Spec.Affinity,
				TopologySpreadConstraints:     d.Agent.Spec.TopologySpreadConstraints,
			},
		},
	}, nil
}
