package operator

import (
	"fmt"
	"path"
	"strings"

	"github.com/grafana/agent/pkg/build"
	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	"github.com/grafana/agent/pkg/operator/clientutil"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func generateLogsDaemonSet(
	cfg *Config,
	name string,
	d gragent.Deployment,
) (*apps_v1.DaemonSet, error) {

	d = *(&d).DeepCopy()

	if d.Agent.Spec.PortName == "" {
		d.Agent.Spec.PortName = defaultPortName
	}

	spec, err := generateLogsDaemonSetSpec(cfg, name, d)
	if err != nil {
		return nil, err
	}

	// Don't transfer any kubectl annotations to the DaemonSet so it doesn't get
	// pruned by kubectl.
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
	labels[agentTypeLabel] = "logs"
	labels[managedByOperatorLabel] = managedByOperatorLabelValue

	ds := &apps_v1.DaemonSet{
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
	// which combines the hash of the DaemonSet, config to the operator, rule
	// config map names (unused here), and the previous DaemonSet (if any).
	//
	// This is used to skip re-applying an unchanged Daemonset. Do we need this?

	if len(d.Agent.Spec.ImagePullSecrets) > 0 {
		ds.Spec.Template.Spec.ImagePullSecrets = d.Agent.Spec.ImagePullSecrets
	}

	return ds, nil
}

func generateLogsDaemonSetSpec(
	cfg *Config,
	name string,
	d gragent.Deployment,
) (*apps_v1.DaemonSetSpec, error) {

	useVersion := d.Agent.Spec.Version
	if useVersion == "" {
		useVersion = DefaultAgentVersion
	}
	imagePath := fmt.Sprintf("%s:%s", DefaultAgentBaseImage, useVersion)
	if d.Agent.Spec.Image != nil && *d.Agent.Spec.Image != "" {
		imagePath = *d.Agent.Spec.Image
	}

	agentArgs := []string{
		"-config.file=/var/lib/grafana-agent/config/agent.yml",
		"-config.expand-env=true",
		"-server.http.address=0.0.0.0:8080",
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
					SecretName: fmt.Sprintf("%s-config", name),
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
		{
			Name: "varlog",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{Path: "/var/log"},
			},
		},
		{
			// Needed for docker. Kubernetes will symlink to this directory. For CRI
			// platforms, this doesn't change anything.
			Name: "dockerlogs",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker/containers"},
			},
		},
		{
			// Needed for storing positions for recovery.
			Name: "data",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/grafana-agent/data"},
			},
		},
	}

	volumeMounts := []v1.VolumeMount{{
		Name:      "config",
		ReadOnly:  true,
		MountPath: "/var/lib/grafana-agent/config-in",
	}, {
		Name:      "config-out",
		MountPath: "/var/lib/grafana-agent/config",
	}, {
		Name:      "secrets",
		ReadOnly:  true,
		MountPath: "/var/lib/grafana-agent/secrets",
	}, {
		Name:      "varlog",
		ReadOnly:  true,
		MountPath: "/var/log",
	}, {
		Name:      "dockerlogs",
		ReadOnly:  true,
		MountPath: "/var/lib/docker/containers",
	}, {
		Name:      "data",
		MountPath: "/var/lib/grafana-agent/data",
	}}
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
			MountPath: path.Join("/var/lib/grafana-agent/extra-secrets", s),
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
			MountPath: path.Join("/var/lib/grafana-agent/extra-configmaps", c),
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
		agentNameLabelName:             d.Agent.Name,
		agentTypeLabel:                 "logs",
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

	envVars := []v1.EnvVar{{
		Name: "POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.name"},
		},
	}, {
		Name: "HOSTNAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
		},
	}, {
		// Not used anywhere for logs but passed to the config-reloader since it
		// expects everything is coming from a StatefulSet.
		Name:  "SHARD",
		Value: "0",
	}}

	operatorContainers := []v1.Container{
		{
			Name:         "config-reloader",
			Image:        "quay.io/prometheus-operator/prometheus-config-reloader:v0.47.0",
			VolumeMounts: volumeMounts,
			Env:          envVars,
			SecurityContext: &v1.SecurityContext{
				Privileged: pointer.Bool(true),
				RunAsUser:  pointer.Int64(0),
			},
			Args: []string{
				"--config-file=/var/lib/grafana-agent/config-in/agent.yml",
				"--config-envsubst-file=/var/lib/grafana-agent/config/agent.yml",

				"--watch-interval=1m",
				"--statefulset-ordinal-from-envvar=SHARD",
				"--reload-url=http://127.0.0.1:8080/-/reload",
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
				ProbeHandler: v1.ProbeHandler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/-/ready",
						Port: intstr.FromString(d.Agent.Spec.PortName),
					},
				},
				InitialDelaySeconds: 10,
				TimeoutSeconds:      probeTimeoutSeconds,
				PeriodSeconds:       5,
			},
			Resources:                d.Agent.Spec.Resources,
			TerminationMessagePolicy: v1.TerminationMessageFallbackToLogsOnError,
		},
	}

	containers, err := clientutil.MergePatchContainers(operatorContainers, d.Agent.Spec.Containers)
	if err != nil {
		return nil, fmt.Errorf("failed to merge containers spec: %w", err)
	}

	return &apps_v1.DaemonSetSpec{
		UpdateStrategy: apps_v1.DaemonSetUpdateStrategy{
			Type: apps_v1.RollingUpdateDaemonSetStrategyType,
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
				TerminationGracePeriodSeconds: pointer.Int64(4800),
				Volumes:                       volumes,
				Tolerations:                   d.Agent.Spec.Tolerations,
				Affinity:                      d.Agent.Spec.Affinity,
				TopologySpreadConstraints:     d.Agent.Spec.TopologySpreadConstraints,
			},
		},
	}, nil
}
