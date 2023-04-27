---
aliases:
- crd/
title: Custom Resource Definition Reference
weight: 500
---
# Custom Resource Definition Reference
<p>Packages:</p>
<ul>
<li>
<a href="#monitoring.grafana.com%2fv1alpha1">monitoring.grafana.com/v1alpha1</a>
</li>
</ul>
<h2 id="monitoring.grafana.com/v1alpha1">monitoring.grafana.com/v1alpha1</h2>
Resource Types:
<ul><li>
<a href="#monitoring.grafana.com/v1alpha1.Deployment">Deployment</a>
</li><li>
<a href="#monitoring.grafana.com/v1alpha1.GrafanaAgent">GrafanaAgent</a>
</li><li>
<a href="#monitoring.grafana.com/v1alpha1.IntegrationsDeployment">IntegrationsDeployment</a>
</li><li>
<a href="#monitoring.grafana.com/v1alpha1.LogsDeployment">LogsDeployment</a>
</li><li>
<a href="#monitoring.grafana.com/v1alpha1.MetricsDeployment">MetricsDeployment</a>
</li></ul>
<h3 id="monitoring.grafana.com/v1alpha1.Deployment">Deployment
</h3>
<div>
<p>Deployment is a set of discovered resources relative to a GrafanaAgent. The
tree of resources contained in a Deployment form the resource hierarchy used
for reconciling a GrafanaAgent.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.grafana.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Deployment</code></td>
</tr>
<tr>
<td>
<code>Agent</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.GrafanaAgent">
GrafanaAgent
</a>
</em>
</td>
<td>
<p>Root resource in the deployment.</p>
</td>
</tr>
<tr>
<td>
<code>Metrics</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MetricsDeployment">
[]MetricsDeployment
</a>
</em>
</td>
<td>
<p>Metrics resources discovered by Agent.</p>
</td>
</tr>
<tr>
<td>
<code>Logs</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsDeployment">
[]LogsDeployment
</a>
</em>
</td>
<td>
<p>Logs resources discovered by Agent.</p>
</td>
</tr>
<tr>
<td>
<code>Integrations</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.IntegrationsDeployment">
[]IntegrationsDeployment
</a>
</em>
</td>
<td>
<p>Integrations resources discovered by Agent.</p>
</td>
</tr>
<tr>
<td>
<code>Secrets</code><br/>
<em>
<a href="https://pkg.go.dev/github.com/grafana/agent/pkg/operator/assets#SecretStore">
github.com/grafana/agent/pkg/operator/assets.SecretStore
</a>
</em>
</td>
<td>
<p>The full list of Secrets referenced by resources in the Deployment.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.GrafanaAgent">GrafanaAgent
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.Deployment">Deployment</a>)
</p>
<div>
<p>GrafanaAgent defines a Grafana Agent deployment.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.grafana.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>GrafanaAgent</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.GrafanaAgentSpec">
GrafanaAgentSpec
</a>
</em>
</td>
<td>
<p>Spec holds the specification of the desired behavior for the Grafana Agent
cluster.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>logLevel</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogLevel controls the log level of the generated pods. Defaults to &ldquo;info&rdquo; if not set.</p>
</td>
</tr>
<tr>
<td>
<code>logFormat</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogFormat controls the logging format of the generated pods. Defaults to &ldquo;logfmt&rdquo; if not set.</p>
</td>
</tr>
<tr>
<td>
<code>apiServer</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.APIServerConfig">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.APIServerConfig
</a>
</em>
</td>
<td>
<p>APIServerConfig allows specifying a host and auth methods to access the
Kubernetes API server. If left empty, the Agent will assume that it is
running inside of the cluster and will discover API servers automatically
and use the pod&rsquo;s CA certificate and bearer token file at
/var/run/secrets/kubernetes.io/serviceaccount.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures Labels and Annotations which are propagated to
created Grafana Agent pods.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of Grafana Agent to be deployed.</p>
</td>
</tr>
<tr>
<td>
<code>paused</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Paused prevents actions except for deletion to be performed on the
underlying managed objects.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image, when specified, overrides the image used to run the Agent. It
should be specified along with a tag. Version must still be set to ensure
the Grafana Agent Operator knows which version of Grafana Agent is being
configured.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>ImagePullSecrets holds an optional list of references to secrets within
the same namespace to use for pulling the Grafana Agent image from
registries.
More info: <a href="https://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.StorageSpec">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.StorageSpec
</a>
</em>
</td>
<td>
<p>Storage spec to specify how storage will be used.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output
StatefulSet definition. Volumes specified will be appended to other
volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output
StatefulSet definition. VolumEMounts specified will be appended to other
VolumeMounts in the Grafana Agent container that are generated as a result
of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Resources holds requests and limits for individual pods.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>NodeSelector defines which nodes pods should be scheduling on.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br/>
<em>
string
</em>
</td>
<td>
<p>ServiceAccountName is the name of the ServiceAccount to use for running Grafana Agent pods.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Secrets is a list of secrets in the same namespace as the GrafanaAgent
object which will be mounted into each running Grafana Agent pod.
The secrets are mounted into /etc/grafana-agent/extra-secrets/<secret-name>.</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>ConfigMaps is a liset of config maps in the same namespace as the
GrafanaAgent object which will be mounted into each running Grafana Agent
pod.
The ConfigMaps are mounted into /etc/grafana-agent/extra-configmaps/<configmap-name>.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>Affinity, if specified, controls pod scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations, if specified, controls the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>TopologySpreadConstraints, if specified, controls the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container
settings. When unspecified, defaults to the default PodSecurityContext.</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers or modifying operator
generated containers. This can be used to allow adding an authentication
proxy to a Grafana Agent pod or to change the behavior of an
operator-generated container. Containers described here modify an operator
generated container if they share the same name and modifications are done
via a strategic merge patch. The current container names are:
<code>grafana-agent</code> and <code>config-reloader</code>. Overriding containers is entirely
outside the scope of what the Grafana Agent team will support and by doing
so, you accept that this behavior may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. These
can be used to, for example, fetch secrets for injection into the Grafana
Agent configuration from external sources. Any errors during the execution
of an initContainer will lead to a restart of the pod.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
Using initContainers for any use case other than secret fetching is
entirely outside the scope of what the Grafana Agent maintainers will
support and by doing so, you accept that this behavior may break at any
time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>priorityClassName</code><br/>
<em>
string
</em>
</td>
<td>
<p>PriorityClassName is the priority class assigned to pods.</p>
</td>
</tr>
<tr>
<td>
<code>portName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Port name used for the pods and governing service. This defaults to agent-metrics.</p>
</td>
</tr>
<tr>
<td>
<code>metrics</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MetricsSubsystemSpec">
MetricsSubsystemSpec
</a>
</em>
</td>
<td>
<p>Metrics controls the metrics subsystem of the Agent and settings
unique to metrics-specific pods that are deployed.</p>
</td>
</tr>
<tr>
<td>
<code>logs</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsSubsystemSpec">
LogsSubsystemSpec
</a>
</em>
</td>
<td>
<p>Logs controls the logging subsystem of the Agent and settings unique to
logging-specific pods that are deployed.</p>
</td>
</tr>
<tr>
<td>
<code>integrations</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.IntegrationsSubsystemSpec">
IntegrationsSubsystemSpec
</a>
</em>
</td>
<td>
<p>Integrations controls the integration subsystem of the Agent and settings
unique to integration-specific pods that are deployed.</p>
</td>
</tr>
<tr>
<td>
<code>enableConfigReadAPI</code><br/>
<em>
bool
</em>
</td>
<td>
<p>enableConfigReadAPI enables the read API for viewing currently running
config port 8080 on the agent.</p>
</td>
</tr>
<tr>
<td>
<code>disableReporting</code><br/>
<em>
bool
</em>
</td>
<td>
<p>disableReporting disable reporting of enabled feature flags to Grafana.</p>
</td>
</tr>
<tr>
<td>
<code>disableSupportBundle</code><br/>
<em>
bool
</em>
</td>
<td>
<p>disableSupportBundle disables the generation of support bundles.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.IntegrationsDeployment">IntegrationsDeployment
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.Deployment">Deployment</a>)
</p>
<div>
<p>IntegrationsDeployment is a set of discovered resources relative to an
IntegrationsDeployment.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.grafana.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>IntegrationsDeployment</code></td>
</tr>
<tr>
<td>
<code>Instance</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.Integration">
Integration
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.LogsDeployment">LogsDeployment
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.Deployment">Deployment</a>)
</p>
<div>
<p>LogsDeployment is a set of discovered resources relative to a LogsInstance.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.grafana.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>LogsDeployment</code></td>
</tr>
<tr>
<td>
<code>Instance</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsInstance">
LogsInstance
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>PodLogs</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.PodLogs">
[]PodLogs
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MetricsDeployment">MetricsDeployment
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.Deployment">Deployment</a>)
</p>
<div>
<p>MetricsDeployment is a set of discovered resources relative to a
MetricsInstance.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code><br/>
string</td>
<td>
<code>
monitoring.grafana.com/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>MetricsDeployment</code></td>
</tr>
<tr>
<td>
<code>Instance</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MetricsInstance">
MetricsInstance
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>ServiceMonitors</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ServiceMonitor">
[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.ServiceMonitor
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>PodMonitors</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PodMonitor">
[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.PodMonitor
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>Probes</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.Probe">
[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.Probe
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.CRIStageSpec">CRIStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>CRIStageSpec is a parsing stage that reads log lines using the standard CRI
logging format. It needs no defined fields.</p>
</div>
<h3 id="monitoring.grafana.com/v1alpha1.DockerStageSpec">DockerStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>DockerStageSpec is a parsing stage that reads log lines using the standard
Docker logging format. It needs no defined fields.</p>
</div>
<h3 id="monitoring.grafana.com/v1alpha1.DropStageSpec">DropStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>DropStageSpec is a filtering stage that lets you drop certain logs.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from the extract data to parse. If empty, uses the log message.</p>
</td>
</tr>
<tr>
<td>
<code>expression</code><br/>
<em>
string
</em>
</td>
<td>
<p>RE2 regular expression.</p>
<p>If source is provided, the regex attempts
to match the source.</p>
<p>If no source is provided, then the regex attempts
to attach the log line.</p>
<p>If the provided regex matches the log line or a provided source, the
line is dropped.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<p>Value can only be specified when source is specified. If the value
provided is an exact match for the given source then the line will be
dropped.</p>
<p>Mutually exclusive with expression.</p>
</td>
</tr>
<tr>
<td>
<code>olderThan</code><br/>
<em>
string
</em>
</td>
<td>
<p>OlderThan will be parsed as a Go duration. If the log line&rsquo;s timestamp
is older than the current time minus the provided duration, it will be
dropped.</p>
</td>
</tr>
<tr>
<td>
<code>longerThan</code><br/>
<em>
string
</em>
</td>
<td>
<p>LongerThan will drop a log line if it its content is longer than this
value (in bytes). Can be expressed as an integer (8192) or a number with a
suffix (8kb).</p>
</td>
</tr>
<tr>
<td>
<code>dropCounterReason</code><br/>
<em>
string
</em>
</td>
<td>
<p>Every time a log line is dropped, the metric logentry_dropped_lines_total
is incremented. A &ldquo;reason&rdquo; label is added, and can be customized by
providing a custom value here. Defaults to &ldquo;drop_stage&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.GrafanaAgentSpec">GrafanaAgentSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.GrafanaAgent">GrafanaAgent</a>)
</p>
<div>
<p>GrafanaAgentSpec is a specification of the desired behavior of the Grafana
Agent cluster.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>logLevel</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogLevel controls the log level of the generated pods. Defaults to &ldquo;info&rdquo; if not set.</p>
</td>
</tr>
<tr>
<td>
<code>logFormat</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogFormat controls the logging format of the generated pods. Defaults to &ldquo;logfmt&rdquo; if not set.</p>
</td>
</tr>
<tr>
<td>
<code>apiServer</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.APIServerConfig">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.APIServerConfig
</a>
</em>
</td>
<td>
<p>APIServerConfig allows specifying a host and auth methods to access the
Kubernetes API server. If left empty, the Agent will assume that it is
running inside of the cluster and will discover API servers automatically
and use the pod&rsquo;s CA certificate and bearer token file at
/var/run/secrets/kubernetes.io/serviceaccount.</p>
</td>
</tr>
<tr>
<td>
<code>podMetadata</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.EmbeddedObjectMetadata">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.EmbeddedObjectMetadata
</a>
</em>
</td>
<td>
<p>PodMetadata configures Labels and Annotations which are propagated to
created Grafana Agent pods.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of Grafana Agent to be deployed.</p>
</td>
</tr>
<tr>
<td>
<code>paused</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Paused prevents actions except for deletion to be performed on the
underlying managed objects.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image, when specified, overrides the image used to run the Agent. It
should be specified along with a tag. Version must still be set to ensure
the Grafana Agent Operator knows which version of Grafana Agent is being
configured.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>ImagePullSecrets holds an optional list of references to secrets within
the same namespace to use for pulling the Grafana Agent image from
registries.
More info: <a href="https://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod</a></p>
</td>
</tr>
<tr>
<td>
<code>storage</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.StorageSpec">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.StorageSpec
</a>
</em>
</td>
<td>
<p>Storage spec to specify how storage will be used.</p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>Volumes allows configuration of additional volumes on the output
StatefulSet definition. Volumes specified will be appended to other
volumes that are generated as a result of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>VolumeMounts allows configuration of additional VolumeMounts on the output
StatefulSet definition. VolumEMounts specified will be appended to other
VolumeMounts in the Grafana Agent container that are generated as a result
of StorageSpec objects.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<p>Resources holds requests and limits for individual pods.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>NodeSelector defines which nodes pods should be scheduling on.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br/>
<em>
string
</em>
</td>
<td>
<p>ServiceAccountName is the name of the ServiceAccount to use for running Grafana Agent pods.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Secrets is a list of secrets in the same namespace as the GrafanaAgent
object which will be mounted into each running Grafana Agent pod.
The secrets are mounted into /etc/grafana-agent/extra-secrets/<secret-name>.</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>ConfigMaps is a liset of config maps in the same namespace as the
GrafanaAgent object which will be mounted into each running Grafana Agent
pod.
The ConfigMaps are mounted into /etc/grafana-agent/extra-configmaps/<configmap-name>.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<p>Affinity, if specified, controls pod scheduling constraints.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations, if specified, controls the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<p>TopologySpreadConstraints, if specified, controls the pod&rsquo;s topology spread constraints.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<p>SecurityContext holds pod-level security attributes and common container
settings. When unspecified, defaults to the default PodSecurityContext.</p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Containers allows injecting additional containers or modifying operator
generated containers. This can be used to allow adding an authentication
proxy to a Grafana Agent pod or to change the behavior of an
operator-generated container. Containers described here modify an operator
generated container if they share the same name and modifications are done
via a strategic merge patch. The current container names are:
<code>grafana-agent</code> and <code>config-reloader</code>. Overriding containers is entirely
outside the scope of what the Grafana Agent team will support and by doing
so, you accept that this behavior may break at any time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>InitContainers allows adding initContainers to the pod definition. These
can be used to, for example, fetch secrets for injection into the Grafana
Agent configuration from external sources. Any errors during the execution
of an initContainer will lead to a restart of the pod.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a>
Using initContainers for any use case other than secret fetching is
entirely outside the scope of what the Grafana Agent maintainers will
support and by doing so, you accept that this behavior may break at any
time without notice.</p>
</td>
</tr>
<tr>
<td>
<code>priorityClassName</code><br/>
<em>
string
</em>
</td>
<td>
<p>PriorityClassName is the priority class assigned to pods.</p>
</td>
</tr>
<tr>
<td>
<code>portName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Port name used for the pods and governing service. This defaults to agent-metrics.</p>
</td>
</tr>
<tr>
<td>
<code>metrics</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MetricsSubsystemSpec">
MetricsSubsystemSpec
</a>
</em>
</td>
<td>
<p>Metrics controls the metrics subsystem of the Agent and settings
unique to metrics-specific pods that are deployed.</p>
</td>
</tr>
<tr>
<td>
<code>logs</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsSubsystemSpec">
LogsSubsystemSpec
</a>
</em>
</td>
<td>
<p>Logs controls the logging subsystem of the Agent and settings unique to
logging-specific pods that are deployed.</p>
</td>
</tr>
<tr>
<td>
<code>integrations</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.IntegrationsSubsystemSpec">
IntegrationsSubsystemSpec
</a>
</em>
</td>
<td>
<p>Integrations controls the integration subsystem of the Agent and settings
unique to integration-specific pods that are deployed.</p>
</td>
</tr>
<tr>
<td>
<code>enableConfigReadAPI</code><br/>
<em>
bool
</em>
</td>
<td>
<p>enableConfigReadAPI enables the read API for viewing currently running
config port 8080 on the agent.</p>
</td>
</tr>
<tr>
<td>
<code>disableReporting</code><br/>
<em>
bool
</em>
</td>
<td>
<p>disableReporting disable reporting of enabled feature flags to Grafana.</p>
</td>
</tr>
<tr>
<td>
<code>disableSupportBundle</code><br/>
<em>
bool
</em>
</td>
<td>
<p>disableSupportBundle disables the generation of support bundles.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.Integration">Integration
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.IntegrationsDeployment">IntegrationsDeployment</a>)
</p>
<div>
<p>Integration runs a single Grafana Agent integration. Integrations that
generate telemetry must be configured to send that telemetry somewhere, such
as autoscrape for exporter-based integrations.</p>
<p>Integrations have access to the LogsInstances and MetricsInstances in the
same GrafanaAgent resource set, referenced by the <namespace>/<name> of the
Instance resource.</p>
<p>For example, if there is a default/production MetricsInstance, you can
configure a supported integration&rsquo;s autoscrape block with:</p>
<pre><code>autoscrape:
enable: true
metrics_instance: default/production
</code></pre>
<p>There is currently no way for telemetry created by an Operator-managed
integration to be collected from outside of the integration itself.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.IntegrationSpec">
IntegrationSpec
</a>
</em>
</td>
<td>
<p>Specifies the desired behavior of the Integration.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the integration to run (e.g., &ldquo;node_exporter&rdquo;, &ldquo;mysqld_exporter&rdquo;).</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.IntegrationType">
IntegrationType
</a>
</em>
</td>
<td>
<p>Type informs Grafana Agent Operator about how to manage the integration being
configured.</p>
</td>
</tr>
<tr>
<td>
<code>config</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<p>The configuration for the named integration. Note that Integrations are
deployed with the integrations-next feature flag, which has different
common settings:</p>
<p><a href="https://grafana.com/docs/agent/latest/configuration/integrations/integrations-next/">https://grafana.com/docs/agent/latest/configuration/integrations/integrations-next/</a></p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>An extra list of Volumes to be associated with the Grafana Agent pods
running this integration. Volume names are mutated to be unique across
all Integrations. Note that the specified volumes should be able to
tolerate existing on multiple pods at once when type is daemonset.</p>
<p>Don&rsquo;t use volumes for loading Secrets or ConfigMaps from the same namespace
as the Integration; use the Secrets and ConfigMaps fields instead.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>An extra list of VolumeMounts to be associated with the Grafana Agent pods
running this integration. VolumeMount names are mutated to be unique
across all used IntegrationSpecs.</p>
<p>Mount paths should include the namespace/name of the Integration CR to
avoid potentially colliding with other resources.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
[]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>An extra list of keys from Secrets in the same namespace as the
Integration which will be mounted into the Grafana Agent pod running this
Integration.</p>
<p>Secrets will be mounted at
/etc/grafana-agent/integrations/secrets/<secret_namespace>/<secret_name>/<key>.</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#configmapkeyselector-v1-core">
[]Kubernetes core/v1.ConfigMapKeySelector
</a>
</em>
</td>
<td>
<p>An extra list of keys from ConfigMaps in the same namespace as the
Integration which will be mounted into the Grafana Agent pod running this
Integration.</p>
<p>ConfigMaps are mounted at
/etc/grafana-agent/integrations/configMaps/<configmap_namespace>/<configmap_name>/<key>.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.IntegrationSpec">IntegrationSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.Integration">Integration</a>)
</p>
<div>
<p>IntegrationSpec specifies the desired behavior of a metrics
integration.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the integration to run (e.g., &ldquo;node_exporter&rdquo;, &ldquo;mysqld_exporter&rdquo;).</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.IntegrationType">
IntegrationType
</a>
</em>
</td>
<td>
<p>Type informs Grafana Agent Operator about how to manage the integration being
configured.</p>
</td>
</tr>
<tr>
<td>
<code>config</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1#JSON">
k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1.JSON
</a>
</em>
</td>
<td>
<p>The configuration for the named integration. Note that Integrations are
deployed with the integrations-next feature flag, which has different
common settings:</p>
<p><a href="https://grafana.com/docs/agent/latest/configuration/integrations/integrations-next/">https://grafana.com/docs/agent/latest/configuration/integrations/integrations-next/</a></p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<p>An extra list of Volumes to be associated with the Grafana Agent pods
running this integration. Volume names are mutated to be unique across
all Integrations. Note that the specified volumes should be able to
tolerate existing on multiple pods at once when type is daemonset.</p>
<p>Don&rsquo;t use volumes for loading Secrets or ConfigMaps from the same namespace
as the Integration; use the Secrets and ConfigMaps fields instead.</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<p>An extra list of VolumeMounts to be associated with the Grafana Agent pods
running this integration. VolumeMount names are mutated to be unique
across all used IntegrationSpecs.</p>
<p>Mount paths should include the namespace/name of the Integration CR to
avoid potentially colliding with other resources.</p>
</td>
</tr>
<tr>
<td>
<code>secrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
[]Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>An extra list of keys from Secrets in the same namespace as the
Integration which will be mounted into the Grafana Agent pod running this
Integration.</p>
<p>Secrets will be mounted at
/etc/grafana-agent/integrations/secrets/<secret_namespace>/<secret_name>/<key>.</p>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#configmapkeyselector-v1-core">
[]Kubernetes core/v1.ConfigMapKeySelector
</a>
</em>
</td>
<td>
<p>An extra list of keys from ConfigMaps in the same namespace as the
Integration which will be mounted into the Grafana Agent pod running this
Integration.</p>
<p>ConfigMaps are mounted at
/etc/grafana-agent/integrations/configMaps/<configmap_namespace>/<configmap_name>/<key>.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.IntegrationType">IntegrationType
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.IntegrationSpec">IntegrationSpec</a>)
</p>
<div>
<p>IntegrationType determines specific behaviors of a configured integration.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>allNodes</code><br/>
<em>
bool
</em>
</td>
<td>
<p>When true, the configured integration should be run on every Node in the
cluster. This is required for Integrations that generate Node-specific
metrics like node_exporter, otherwise it must be false to avoid generating
duplicate metrics.</p>
</td>
</tr>
<tr>
<td>
<code>unique</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Whether this integration can only be defined once for a Grafana Agent
process, such as statsd_exporter. It is invalid for a GrafanaAgent to
discover multiple unique Integrations with the same Integration name
(i.e., a single GrafanaAgent cannot deploy two statsd_exporters).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.IntegrationsSubsystemSpec">IntegrationsSubsystemSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.GrafanaAgentSpec">GrafanaAgentSpec</a>)
</p>
<div>
<p>IntegrationsSubsystemSpec defines global settings to apply across the
integrations subsystem.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Label selector to find Integration resources to run. When nil, no
integration resources will be defined.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Label selector for namespaces to search when discovering integration
resources. If nil, integration resources are only discovered in the
namespace of the GrafanaAgent resource.</p>
<p>Set to <code>{}</code> to search all namespaces.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.JSONStageSpec">JSONStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>JSONStageSpec is a parsing stage that reads the log line as JSON and accepts
JMESPath expressions to extract data.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from the extracted data to parse as JSON. If empty, uses entire log
message.</p>
</td>
</tr>
<tr>
<td>
<code>expressions</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Set of the key/value pairs of JMESPath expressions. The key will be the
key in the extracted data while the expression will be the value,
evaluated as a JMESPath from the source data.</p>
<p>Literal JMESPath expressions can be used by wrapping a key in double
quotes, which then must be wrapped again in single quotes in YAML
so they get passed to the JMESPath parser.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.LogsBackoffConfigSpec">LogsBackoffConfigSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.LogsClientSpec">LogsClientSpec</a>)
</p>
<div>
<p>LogsBackoffConfigSpec configures timing for retrying failed requests.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>minPeriod</code><br/>
<em>
string
</em>
</td>
<td>
<p>Initial backoff time between retries. Time between retries is
increased exponentially.</p>
</td>
</tr>
<tr>
<td>
<code>maxPeriod</code><br/>
<em>
string
</em>
</td>
<td>
<p>Maximum backoff time between retries.</p>
</td>
</tr>
<tr>
<td>
<code>maxRetries</code><br/>
<em>
int
</em>
</td>
<td>
<p>Maximum number of retries to perform before giving up a request.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.LogsClientSpec">LogsClientSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.LogsInstanceSpec">LogsInstanceSpec</a>, <a href="#monitoring.grafana.com/v1alpha1.LogsSubsystemSpec">LogsSubsystemSpec</a>)
</p>
<div>
<p>LogsClientSpec defines the client integration for logs, indicating which
Loki server to send logs to.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<p>URL is the URL where Loki is listening. Must be a full HTTP URL, including
protocol. Required.
Example: <a href="https://logs-prod-us-central1.grafana.net/loki/api/v1/push">https://logs-prod-us-central1.grafana.net/loki/api/v1/push</a>.</p>
</td>
</tr>
<tr>
<td>
<code>tenantId</code><br/>
<em>
string
</em>
</td>
<td>
<p>Tenant ID used by default to push logs to Loki. If omitted assumes remote
Loki is running in single-tenant mode or an authentication layer is used
to inject an X-Scope-OrgID header.</p>
</td>
</tr>
<tr>
<td>
<code>batchWait</code><br/>
<em>
string
</em>
</td>
<td>
<p>Maximum amount of time to wait before sending a batch, even if that batch
isn&rsquo;t full.</p>
</td>
</tr>
<tr>
<td>
<code>batchSize</code><br/>
<em>
int
</em>
</td>
<td>
<p>Maximum batch size (in bytes) of logs to accumulate before sending the
batch to Loki.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.BasicAuth">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth for the Loki server.</p>
</td>
</tr>
<tr>
<td>
<code>bearerToken</code><br/>
<em>
string
</em>
</td>
<td>
<p>BearerToken used for remote_write.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>BearerTokenFile used to read bearer token.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProxyURL to proxy requests through. Optional.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.TLSConfig">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.TLSConfig
</a>
</em>
</td>
<td>
<p>TLSConfig to use for the client. Only used when the protocol of the URL
is https.</p>
</td>
</tr>
<tr>
<td>
<code>backoffConfig</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsBackoffConfigSpec">
LogsBackoffConfigSpec
</a>
</em>
</td>
<td>
<p>Configures how to retry requests to Loki when a request fails.
Defaults to a minPeriod of 500ms, maxPeriod of 5m, and maxRetries of 10.</p>
</td>
</tr>
<tr>
<td>
<code>externalLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>ExternalLabels are labels to add to any time series when sending data to
Loki.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br/>
<em>
string
</em>
</td>
<td>
<p>Maximum time to wait for a server to respond to a request.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.LogsInstance">LogsInstance
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.LogsDeployment">LogsDeployment</a>)
</p>
<div>
<p>LogsInstance controls an individual logs instance within a Grafana Agent
deployment.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsInstanceSpec">
LogsInstanceSpec
</a>
</em>
</td>
<td>
<p>Spec holds the specification of the desired behavior for the logs
instance.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>clients</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsClientSpec">
[]LogsClientSpec
</a>
</em>
</td>
<td>
<p>Clients controls where logs are written to for this instance.</p>
</td>
</tr>
<tr>
<td>
<code>podLogsSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Determines which PodLogs should be selected for including in this
instance.</p>
</td>
</tr>
<tr>
<td>
<code>podLogsNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Set of labels to determine which namespaces should be watched
for PodLogs. If not provided, checks only namespace of the
instance.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalScrapeConfigs allows specifying a key of a Secret containing
additional Grafana Agent logging scrape configurations. Scrape
configurations specified are appended to the configurations generated by
the Grafana Agent Operator.</p>
<p>Job configurations specified must have the form as specified in the
official Promtail documentation:</p>
<p><a href="https://grafana.com/docs/loki/latest/clients/promtail/configuration/#scrape_configs">https://grafana.com/docs/loki/latest/clients/promtail/configuration/#scrape_configs</a></p>
<p>As scrape configs are appended, the user is responsible to make sure it is
valid. Note that using this feature may expose the possibility to break
upgrades of Grafana Agent. It is advised to review both Grafana Agent and
Promtail release notes to ensure that no incompatible scrape configs are
going to break Grafana Agent after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>targetConfig</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsTargetConfigSpec">
LogsTargetConfigSpec
</a>
</em>
</td>
<td>
<p>Configures how tailed targets are watched.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.LogsInstanceSpec">LogsInstanceSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.LogsInstance">LogsInstance</a>)
</p>
<div>
<p>LogsInstanceSpec controls how an individual instance will be used to
discover LogMonitors.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>clients</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsClientSpec">
[]LogsClientSpec
</a>
</em>
</td>
<td>
<p>Clients controls where logs are written to for this instance.</p>
</td>
</tr>
<tr>
<td>
<code>podLogsSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Determines which PodLogs should be selected for including in this
instance.</p>
</td>
</tr>
<tr>
<td>
<code>podLogsNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Set of labels to determine which namespaces should be watched
for PodLogs. If not provided, checks only namespace of the
instance.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalScrapeConfigs allows specifying a key of a Secret containing
additional Grafana Agent logging scrape configurations. Scrape
configurations specified are appended to the configurations generated by
the Grafana Agent Operator.</p>
<p>Job configurations specified must have the form as specified in the
official Promtail documentation:</p>
<p><a href="https://grafana.com/docs/loki/latest/clients/promtail/configuration/#scrape_configs">https://grafana.com/docs/loki/latest/clients/promtail/configuration/#scrape_configs</a></p>
<p>As scrape configs are appended, the user is responsible to make sure it is
valid. Note that using this feature may expose the possibility to break
upgrades of Grafana Agent. It is advised to review both Grafana Agent and
Promtail release notes to ensure that no incompatible scrape configs are
going to break Grafana Agent after the upgrade.</p>
</td>
</tr>
<tr>
<td>
<code>targetConfig</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsTargetConfigSpec">
LogsTargetConfigSpec
</a>
</em>
</td>
<td>
<p>Configures how tailed targets are watched.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.LogsSubsystemSpec">LogsSubsystemSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.GrafanaAgentSpec">GrafanaAgentSpec</a>)
</p>
<div>
<p>LogsSubsystemSpec defines global settings to apply across the logging
subsystem.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>clients</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.LogsClientSpec">
[]LogsClientSpec
</a>
</em>
</td>
<td>
<p>A global set of clients to use when a discovered LogsInstance does not
have any clients defined.</p>
</td>
</tr>
<tr>
<td>
<code>logsExternalLabelName</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogsExternalLabelName is the name of the external label used to
denote Grafana Agent cluster. Defaults to &ldquo;cluster.&rdquo; External label will
<em>not</em> be added when value is set to the empty string.</p>
</td>
</tr>
<tr>
<td>
<code>instanceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>InstanceSelector determines which LogInstances should be selected
for running. Each instance runs its own set of Prometheus components,
including service discovery, scraping, and remote_write.</p>
</td>
</tr>
<tr>
<td>
<code>instanceNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>InstanceNamespaceSelector are the set of labels to determine which
namespaces to watch for LogInstances. If not provided, only checks own
namespace.</p>
</td>
</tr>
<tr>
<td>
<code>ignoreNamespaceSelectors</code><br/>
<em>
bool
</em>
</td>
<td>
<p>IgnoreNamespaceSelectors, if true, will ignore NamespaceSelector settings
from the PodLogs configs, and they will only discover endpoints within
their current namespace.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedNamespaceLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>EnforcedNamespaceLabel enforces adding a namespace label of origin for
each metric that is user-created. The label value will always be the
namespace of the object that is being created.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.LogsTargetConfigSpec">LogsTargetConfigSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.LogsInstanceSpec">LogsInstanceSpec</a>)
</p>
<div>
<p>LogsTargetConfigSpec configures how tailed targets are watched.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>syncPeriod</code><br/>
<em>
string
</em>
</td>
<td>
<p>Period to resync directories being watched and files being tailed to discover
new ones or stop watching removed ones.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MatchStageSpec">MatchStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>MatchStageSpec is a filtering stage that conditionally applies a set of
stages or drop entries when a log entry matches a configurable LogQL stream
selector and filter expressions.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>selector</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogQL stream selector and filter expressions. Required.</p>
</td>
</tr>
<tr>
<td>
<code>pipelineName</code><br/>
<em>
string
</em>
</td>
<td>
<p>Names the pipeline. When defined, creates an additional label
in the pipeline_duration_seconds histogram, where the value is
concatenated with job_name using an underscore.</p>
</td>
</tr>
<tr>
<td>
<code>action</code><br/>
<em>
string
</em>
</td>
<td>
<p>Determines what action is taken when the selector matches the log line.
Can be keep or drop. Defaults to keep. When set to drop, entries are
dropped and no later metrics are recorded.
Stages must be empty when dropping metrics.</p>
</td>
</tr>
<tr>
<td>
<code>dropCounterReason</code><br/>
<em>
string
</em>
</td>
<td>
<p>Every time a log line is dropped, the metric logentry_dropped_lines_total
is incremented. A &ldquo;reason&rdquo; label is added, and can be customized by
providing a custom value here. Defaults to &ldquo;match_stage.&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>stages</code><br/>
<em>
string
</em>
</td>
<td>
<p>Nested set of pipeline stages to execute when action is keep and the log
line matches selector.</p>
<p>An example value for stages may be:</p>
<p>stages: |
- json: {}
- labelAllow: [foo, bar]</p>
<p>Note that stages is a string because SIG API Machinery does not
support recursive types, and so it cannot be validated for correctness. Be
careful not to mistype anything.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MetadataConfig">MetadataConfig
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>MetadataConfig configures the sending of series metadata to remote storage.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>send</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Send enables metric metadata to be sent to remote storage.</p>
</td>
</tr>
<tr>
<td>
<code>sendInterval</code><br/>
<em>
string
</em>
</td>
<td>
<p>SendInterval controls how frequently metric metadata is sent to remote storage.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MetricsInstance">MetricsInstance
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.MetricsDeployment">MetricsDeployment</a>)
</p>
<div>
<p>MetricsInstance controls an individual Metrics instance within a
Grafana Agent deployment.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MetricsInstanceSpec">
MetricsInstanceSpec
</a>
</em>
</td>
<td>
<p>Spec holds the specification of the desired behavior for the Metrics
instance.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>walTruncateFrequency</code><br/>
<em>
string
</em>
</td>
<td>
<p>WALTruncateFrequency specifies how frequently the WAL truncation process
should run. Higher values causes the WAL to increase and for old series to
stay in the WAL for longer, but reduces the chances of data loss when
remote_write is failing for longer than the given frequency.</p>
</td>
</tr>
<tr>
<td>
<code>minWALTime</code><br/>
<em>
string
</em>
</td>
<td>
<p>MinWALTime is the minimum amount of time series and samples may exist in
the WAL before being considered for deletion.</p>
</td>
</tr>
<tr>
<td>
<code>maxWALTime</code><br/>
<em>
string
</em>
</td>
<td>
<p>MaxWALTime is the maximum amount of time series and asmples may exist in
the WAL before being forcibly deleted.</p>
</td>
</tr>
<tr>
<td>
<code>remoteFlushDeadline</code><br/>
<em>
string
</em>
</td>
<td>
<p>RemoteFlushDeadline is the deadline for flushing data when an instance
shuts down.</p>
</td>
</tr>
<tr>
<td>
<code>writeStaleOnShutdown</code><br/>
<em>
bool
</em>
</td>
<td>
<p>WriteStaleOnShutdown writes staleness markers on shutdown for all series.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitorSelector determines which ServiceMonitors should be selected
for target discovery.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitorNamespaceSelector are the set of labels to determine which
namespaces to watch for ServiceMonitor discovery. If nil, only checks own
namespace.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PodMonitorSelector determines which PodMonitors should be selected for target
discovery. Experimental.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PodMonitorNamespaceSelector are the set of labels to determine which
namespaces to watch for PodMonitor discovery. If nil, only checks own
namespace.</p>
</td>
</tr>
<tr>
<td>
<code>probeSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ProbeSelector determines which Probes should be selected for target
discovery.</p>
</td>
</tr>
<tr>
<td>
<code>probeNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ProbeNamespaceSelector are the set of labels to determine which namespaces
to watch for Probe discovery. If nil, only checks own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.RemoteWriteSpec">
[]RemoteWriteSpec
</a>
</em>
</td>
<td>
<p>RemoteWrite controls remote_write settings for this instance.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalScrapeConfigs allows specifying a key of a Secret containing
additional Grafana Agent Prometheus scrape configurations. SCrape
configurations specified are appended to the configurations generated by
the Grafana Agent Operator. Job configurations specified must have the
form as specified in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config</a>.
As scrape configs are appended, the user is responsible to make sure it is
valid. Note that using this feature may expose the possibility to break
upgrades of Grafana Agent. It is advised to review both Grafana Agent and
Prometheus release notes to ensure that no incompatible scrape configs are
going to break Grafana Agent after the upgrade.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MetricsInstanceSpec">MetricsInstanceSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.MetricsInstance">MetricsInstance</a>)
</p>
<div>
<p>MetricsInstanceSpec controls how an individual instance will be used to discover PodMonitors.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>walTruncateFrequency</code><br/>
<em>
string
</em>
</td>
<td>
<p>WALTruncateFrequency specifies how frequently the WAL truncation process
should run. Higher values causes the WAL to increase and for old series to
stay in the WAL for longer, but reduces the chances of data loss when
remote_write is failing for longer than the given frequency.</p>
</td>
</tr>
<tr>
<td>
<code>minWALTime</code><br/>
<em>
string
</em>
</td>
<td>
<p>MinWALTime is the minimum amount of time series and samples may exist in
the WAL before being considered for deletion.</p>
</td>
</tr>
<tr>
<td>
<code>maxWALTime</code><br/>
<em>
string
</em>
</td>
<td>
<p>MaxWALTime is the maximum amount of time series and asmples may exist in
the WAL before being forcibly deleted.</p>
</td>
</tr>
<tr>
<td>
<code>remoteFlushDeadline</code><br/>
<em>
string
</em>
</td>
<td>
<p>RemoteFlushDeadline is the deadline for flushing data when an instance
shuts down.</p>
</td>
</tr>
<tr>
<td>
<code>writeStaleOnShutdown</code><br/>
<em>
bool
</em>
</td>
<td>
<p>WriteStaleOnShutdown writes staleness markers on shutdown for all series.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitorSelector determines which ServiceMonitors should be selected
for target discovery.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ServiceMonitorNamespaceSelector are the set of labels to determine which
namespaces to watch for ServiceMonitor discovery. If nil, only checks own
namespace.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PodMonitorSelector determines which PodMonitors should be selected for target
discovery. Experimental.</p>
</td>
</tr>
<tr>
<td>
<code>podMonitorNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PodMonitorNamespaceSelector are the set of labels to determine which
namespaces to watch for PodMonitor discovery. If nil, only checks own
namespace.</p>
</td>
</tr>
<tr>
<td>
<code>probeSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ProbeSelector determines which Probes should be selected for target
discovery.</p>
</td>
</tr>
<tr>
<td>
<code>probeNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ProbeNamespaceSelector are the set of labels to determine which namespaces
to watch for Probe discovery. If nil, only checks own namespace.</p>
</td>
</tr>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.RemoteWriteSpec">
[]RemoteWriteSpec
</a>
</em>
</td>
<td>
<p>RemoteWrite controls remote_write settings for this instance.</p>
</td>
</tr>
<tr>
<td>
<code>additionalScrapeConfigs</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AdditionalScrapeConfigs allows specifying a key of a Secret containing
additional Grafana Agent Prometheus scrape configurations. SCrape
configurations specified are appended to the configurations generated by
the Grafana Agent Operator. Job configurations specified must have the
form as specified in the official Prometheus documentation:
<a href="https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config">https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config</a>.
As scrape configs are appended, the user is responsible to make sure it is
valid. Note that using this feature may expose the possibility to break
upgrades of Grafana Agent. It is advised to review both Grafana Agent and
Prometheus release notes to ensure that no incompatible scrape configs are
going to break Grafana Agent after the upgrade.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MetricsStageSpec">MetricsStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>MetricsStageSpec is an action stage that allows for defining and updating
metrics based on data from the extracted map. Created metrics are not pushed
to Loki or Prometheus and are instead exposed via the /metrics endpoint of
the Grafana Agent pod. The Grafana Agent Operator should be configured with
a MetricsInstance that discovers the logging DaemonSet to collect metrics
created by this stage.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br/>
<em>
string
</em>
</td>
<td>
<p>The metric type to create. Must be one of counter, gauge, histogram.
Required.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Sets the description for the created metric.</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br/>
<em>
string
</em>
</td>
<td>
<p>Sets the custom prefix name for the metric. Defaults to &ldquo;promtail<em>custom</em>&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Key from the extracted data map to use for the metric. Defaults to the
metrics name if not present.</p>
</td>
</tr>
<tr>
<td>
<code>maxIdleDuration</code><br/>
<em>
string
</em>
</td>
<td>
<p>Label values on metrics are dynamic which can cause exported metrics
to go stale. To prevent unbounded cardinality, any metrics not updated
within MaxIdleDuration are removed.</p>
<p>Must be greater or equal to 1s. Defaults to 5m.</p>
</td>
</tr>
<tr>
<td>
<code>matchAll</code><br/>
<em>
bool
</em>
</td>
<td>
<p>If true, all log lines are counted without attempting to match the
source to the extracted map. Mutually exclusive with value.</p>
<p>Only valid for type: counter.</p>
</td>
</tr>
<tr>
<td>
<code>countEntryBytes</code><br/>
<em>
bool
</em>
</td>
<td>
<p>If true all log line bytes are counted. Can only be set with
matchAll: true and action: add.</p>
<p>Only valid for type: counter.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<p>Filters down source data and only changes the metric if the targeted
value matches the provided string exactly. If not present, all
data matches.</p>
</td>
</tr>
<tr>
<td>
<code>action</code><br/>
<em>
string
</em>
</td>
<td>
<p>The action to take against the metric. Required.</p>
<p>Must be either &ldquo;inc&rdquo; or &ldquo;add&rdquo; for type: counter or type: histogram.
When type: gauge, must be one of &ldquo;set&rdquo;, &ldquo;inc&rdquo;, &ldquo;dec&rdquo;, &ldquo;add&rdquo;, or &ldquo;sub&rdquo;.</p>
<p>&ldquo;add&rdquo;, &ldquo;set&rdquo;, or &ldquo;sub&rdquo; requires the extracted value to be convertible
to a positive float.</p>
</td>
</tr>
<tr>
<td>
<code>buckets</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Buckets to create. Bucket values must be convertible to float64s. Extremely
large or small numbers are subject to some loss of precision.
Only valid for type: histogram.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MetricsSubsystemSpec">MetricsSubsystemSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.GrafanaAgentSpec">GrafanaAgentSpec</a>)
</p>
<div>
<p>MetricsSubsystemSpec defines global settings to apply across the
Metrics subsystem.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>remoteWrite</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.RemoteWriteSpec">
[]RemoteWriteSpec
</a>
</em>
</td>
<td>
<p>RemoteWrite controls default remote_write settings for all instances. If
an instance does not provide its own remoteWrite settings, these will be
used instead.</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Replicas of each shard to deploy for metrics pods. Number of replicas
multiplied by the number of shards is the total number of pods created.</p>
</td>
</tr>
<tr>
<td>
<code>shards</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Shards to distribute targets onto. Number of replicas multiplied by the
number of shards is the total number of pods created. Note that scaling
down shards will not reshard data onto remaining instances, it must be
manually moved. Increasing shards will not reshard data either but it will
continue to be available from the same instances. Sharding is performed on
the content of the <strong>address</strong> target meta-label.</p>
</td>
</tr>
<tr>
<td>
<code>replicaExternalLabelName</code><br/>
<em>
string
</em>
</td>
<td>
<p>ReplicaExternalLabelName is the name of the metrics external label used
to denote replica name. Defaults to <strong>replica</strong>. External label will <em>not</em>
be added when value is set to the empty string.</p>
</td>
</tr>
<tr>
<td>
<code>metricsExternalLabelName</code><br/>
<em>
string
</em>
</td>
<td>
<p>MetricsExternalLabelName is the name of the external label used to
denote Grafana Agent cluster. Defaults to &ldquo;cluster.&rdquo; External label will
<em>not</em> be added when value is set to the empty string.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeInterval</code><br/>
<em>
string
</em>
</td>
<td>
<p>ScrapeInterval is the time between consecutive scrapes.</p>
</td>
</tr>
<tr>
<td>
<code>scrapeTimeout</code><br/>
<em>
string
</em>
</td>
<td>
<p>ScrapeTimeout is the time to wait for a target to respond before marking a
scrape as failed.</p>
</td>
</tr>
<tr>
<td>
<code>externalLabels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>ExternalLabels are labels to add to any time series when sending data over
remote_write.</p>
</td>
</tr>
<tr>
<td>
<code>arbitraryFSAccessThroughSMs</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ArbitraryFSAccessThroughSMsConfig">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.ArbitraryFSAccessThroughSMsConfig
</a>
</em>
</td>
<td>
<p>ArbitraryFSAccessThroughSMs configures whether configuration based on a
ServiceMonitor can access arbitrary files on the file system of the
Grafana Agent container e.g. bearer token files.</p>
</td>
</tr>
<tr>
<td>
<code>overrideHonorLabels</code><br/>
<em>
bool
</em>
</td>
<td>
<p>OverrideHonorLabels, if true, overrides all configured honor_labels read
from ServiceMonitor or PodMonitor to false.</p>
</td>
</tr>
<tr>
<td>
<code>overrideHonorTimestamps</code><br/>
<em>
bool
</em>
</td>
<td>
<p>OverrideHonorTimestamps allows to globally enforce honoring timestamps in all scrape configs.</p>
</td>
</tr>
<tr>
<td>
<code>ignoreNamespaceSelectors</code><br/>
<em>
bool
</em>
</td>
<td>
<p>IgnoreNamespaceSelectors, if true, will ignore NamespaceSelector settings
from the PodMonitor and ServiceMonitor configs, and they will only
discover endpoints within their current namespace.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedNamespaceLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>EnforcedNamespaceLabel enforces adding a namespace label of origin for
each metric that is user-created. The label value will always be the
namespace of the object that is being created.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedSampleLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<p>EnforcedSampleLimit defines global limit on the number of scraped samples
that will be accepted. This overrides any SampleLimit set per
ServiceMonitor and/or PodMonitor. It is meant to be used by admins to
enforce the SampleLimit to keep the overall number of samples and series
under the desired limit. Note that if a SampleLimit from a ServiceMonitor
or PodMonitor is lower, that value will be used instead.</p>
</td>
</tr>
<tr>
<td>
<code>enforcedTargetLimit</code><br/>
<em>
uint64
</em>
</td>
<td>
<p>EnforcedTargetLimit defines a global limit on the number of scraped
targets. This overrides any TargetLimit set per ServiceMonitor and/or
PodMonitor. It is meant to be used by admins to enforce the TargetLimit to
keep the overall number of targets under the desired limit. Note that if a
TargetLimit from a ServiceMonitor or PodMonitor is higher, that value will
be used instead.</p>
</td>
</tr>
<tr>
<td>
<code>instanceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>InstanceSelector determines which MetricsInstances should be selected
for running. Each instance runs its own set of Metrics components,
including service discovery, scraping, and remote_write.</p>
</td>
</tr>
<tr>
<td>
<code>instanceNamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>InstanceNamespaceSelector are the set of labels to determine which
namespaces to watch for MetricsInstances. If not provided, only checks own namespace.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.MultilineStageSpec">MultilineStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>MultilineStageSpec merges multiple lines into a multiline block before
passing it on to the next stage in the pipeline.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>firstLine</code><br/>
<em>
string
</em>
</td>
<td>
<p>RE2 regular expression. Creates a new multiline block when matched.
Required.</p>
</td>
</tr>
<tr>
<td>
<code>maxWaitTime</code><br/>
<em>
string
</em>
</td>
<td>
<p>Maximum time to wait before passing on the multiline block to the next
stage if no new lines are received. Defaults to 3s.</p>
</td>
</tr>
<tr>
<td>
<code>maxLines</code><br/>
<em>
int
</em>
</td>
<td>
<p>Maximum number of lines a block can have. A new block is started if
the number of lines surpasses this value. Defaults to 128.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.ObjectSelector">ObjectSelector
</h3>
<div>
<p>ObjectSelector is a set of selectors to use for finding an object in the
resource hierarchy. When NamespaceSelector is nil, objects should be
searched directly in the ParentNamespace.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ObjectType</code><br/>
<em>
<a href="https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#Object">
sigs.k8s.io/controller-runtime/pkg/client.Object
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>ParentNamespace</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>NamespaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>Labels</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.OutputStageSpec">OutputStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>OutputStageSpec is an action stage that takes data from the extracted map
and changes the log line that will be sent to Loki.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from extract data to use for the log entry. Required.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.PackStageSpec">PackStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>PackStageSpec is a transform stage that lets you embed extracted values and
labels into the log line by packing the log line and labels inside of a JSON
object.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>labels</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Name from extracted data or line labels. Required.
Labels provided here are automatically removed from output labels.</p>
</td>
</tr>
<tr>
<td>
<code>ingestTimestamp</code><br/>
<em>
bool
</em>
</td>
<td>
<p>If the resulting log line should use any existing timestamp or use time.Now()
when the line was created. Set to true when combining several log streams from
different containers to avoid out of order errors.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PodLogsSpec">PodLogsSpec</a>)
</p>
<div>
<p>PipelineStageSpec defines an individual pipeline stage. Each stage type is
mutually exclusive and no more than one may be set per stage.</p>
<p>More information on pipelines can be found in the Promtail documentation:
<a href="https://grafana.com/docs/loki/latest/clients/promtail/pipelines/">https://grafana.com/docs/loki/latest/clients/promtail/pipelines/</a></p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>cri</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.CRIStageSpec">
CRIStageSpec
</a>
</em>
</td>
<td>
<p>CRI is a parsing stage that reads log lines using the standard
CRI logging format. Supply cri: {} to enable.</p>
</td>
</tr>
<tr>
<td>
<code>docker</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.DockerStageSpec">
DockerStageSpec
</a>
</em>
</td>
<td>
<p>Docker is a parsing stage that reads log lines using the standard
Docker logging format. Supply docker: {} to enable.</p>
</td>
</tr>
<tr>
<td>
<code>drop</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.DropStageSpec">
DropStageSpec
</a>
</em>
</td>
<td>
<p>Drop is a filtering stage that lets you drop certain logs.</p>
</td>
</tr>
<tr>
<td>
<code>json</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.JSONStageSpec">
JSONStageSpec
</a>
</em>
</td>
<td>
<p>JSON is a parsing stage that reads the log line as JSON and accepts
JMESPath expressions to extract data.</p>
<p>Information on JMESPath: <a href="http://jmespath.org/">http://jmespath.org/</a></p>
</td>
</tr>
<tr>
<td>
<code>labelAllow</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>LabelAllow is an action stage that only allows the provided labels to be
included in the label set that is sent to Loki with the log entry.</p>
</td>
</tr>
<tr>
<td>
<code>labelDrop</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>LabelDrop is an action stage that drops labels from the label set that
is sent to Loki with the log entry.</p>
</td>
</tr>
<tr>
<td>
<code>labels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Labels is an action stage that takes data from the extracted map and
modifies the label set that is sent to Loki with the log entry.</p>
<p>The key is REQUIRED and represents the name for the label that will
be created. Value is optional and will be the name from extracted data
to use for the value of the label. If the value is not provided, it
defaults to match the key.</p>
</td>
</tr>
<tr>
<td>
<code>match</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MatchStageSpec">
MatchStageSpec
</a>
</em>
</td>
<td>
<p>Match is a filtering stage that conditionally applies a set of stages
or drop entries when a log entry matches a configurable LogQL stream
selector and filter expressions.</p>
</td>
</tr>
<tr>
<td>
<code>metrics</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MetricsStageSpec">
map[string]github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1.MetricsStageSpec
</a>
</em>
</td>
<td>
<p>Metrics is an action stage that supports defining and updating metrics
based on data from the extracted map. Created metrics are not pushed to
Loki or Prometheus and are instead exposed via the /metrics endpoint of
the Grafana Agent pod. The Grafana Agent Operator should be configured
with a MetricsInstance that discovers the logging DaemonSet to collect
metrics created by this stage.</p>
</td>
</tr>
<tr>
<td>
<code>multiline</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MultilineStageSpec">
MultilineStageSpec
</a>
</em>
</td>
<td>
<p>Multiline stage merges multiple lines into a multiline block before
passing it on to the next stage in the pipeline.</p>
</td>
</tr>
<tr>
<td>
<code>output</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.OutputStageSpec">
OutputStageSpec
</a>
</em>
</td>
<td>
<p>Output stage is an action stage that takes data from the extracted map and
changes the log line that will be sent to Loki.</p>
</td>
</tr>
<tr>
<td>
<code>pack</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.PackStageSpec">
PackStageSpec
</a>
</em>
</td>
<td>
<p>Pack is a transform stage that lets you embed extracted values and labels
into the log line by packing the log line and labels inside of a JSON
object.</p>
</td>
</tr>
<tr>
<td>
<code>regex</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.RegexStageSpec">
RegexStageSpec
</a>
</em>
</td>
<td>
<p>Regex is a parsing stage that parses a log line using a regular
expression.  Named capture groups in the regex allows for adding data into
the extracted map.</p>
</td>
</tr>
<tr>
<td>
<code>replace</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.ReplaceStageSpec">
ReplaceStageSpec
</a>
</em>
</td>
<td>
<p>Replace is a parsing stage that parses a log line using a regular
expression and replaces the log line. Named capture groups in the regex
allows for adding data into the extracted map.</p>
</td>
</tr>
<tr>
<td>
<code>template</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.TemplateStageSpec">
TemplateStageSpec
</a>
</em>
</td>
<td>
<p>Template is a transform stage that manipulates the values in the extracted
map using Go&rsquo;s template syntax.</p>
</td>
</tr>
<tr>
<td>
<code>tenant</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.TenantStageSpec">
TenantStageSpec
</a>
</em>
</td>
<td>
<p>Tenant is an action stage that sets the tenant ID for the log entry picking it from a
field in the extracted data map. If the field is missing, the default
LogsClientSpec.tenantId will be used.</p>
</td>
</tr>
<tr>
<td>
<code>timestamp</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.TimestampStageSpec">
TimestampStageSpec
</a>
</em>
</td>
<td>
<p>Timestamp is an action stage that can change the timestamp of a log line
before it is sent to Loki. If not present, the timestamp of a log line
defaults to the time when the log line was read.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.PodLogs">PodLogs
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.LogsDeployment">LogsDeployment</a>)
</p>
<div>
<p>PodLogs defines how to collect logs for a pod.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.PodLogsSpec">
PodLogsSpec
</a>
</em>
</td>
<td>
<p>Spec holds the specification of the desired behavior for the PodLogs.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>jobLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>The label to use to retrieve the job name from.</p>
</td>
</tr>
<tr>
<td>
<code>podTargetLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>PodTargetLabels transfers labels on the Kubernetes Pod onto the target.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector to select Pod objects. Required.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.NamespaceSelector">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.NamespaceSelector
</a>
</em>
</td>
<td>
<p>Selector to select which namespaces the Pod objects are discovered from.</p>
</td>
</tr>
<tr>
<td>
<code>pipelineStages</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">
[]PipelineStageSpec
</a>
</em>
</td>
<td>
<p>Pipeline stages for this pod. Pipeline stages support transforming and
filtering log lines.</p>
</td>
</tr>
<tr>
<td>
<code>relabelings</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.RelabelConfig">
[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.RelabelConfig
</a>
</em>
</td>
<td>
<p>RelabelConfigs to apply to logs before delivering.
Grafana Agent Operator automatically adds relabelings for a few standard
Kubernetes fields and replaces original scrape job name with
__tmp_logs_job_name.</p>
<p>More info: <a href="https://grafana.com/docs/loki/latest/clients/promtail/configuration/#relabel_configs">https://grafana.com/docs/loki/latest/clients/promtail/configuration/#relabel_configs</a></p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.PodLogsSpec">PodLogsSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PodLogs">PodLogs</a>)
</p>
<div>
<p>PodLogsSpec defines how to collect logs for a pod.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>jobLabel</code><br/>
<em>
string
</em>
</td>
<td>
<p>The label to use to retrieve the job name from.</p>
</td>
</tr>
<tr>
<td>
<code>podTargetLabels</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>PodTargetLabels transfers labels on the Kubernetes Pod onto the target.</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector to select Pod objects. Required.</p>
</td>
</tr>
<tr>
<td>
<code>namespaceSelector</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.NamespaceSelector">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.NamespaceSelector
</a>
</em>
</td>
<td>
<p>Selector to select which namespaces the Pod objects are discovered from.</p>
</td>
</tr>
<tr>
<td>
<code>pipelineStages</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">
[]PipelineStageSpec
</a>
</em>
</td>
<td>
<p>Pipeline stages for this pod. Pipeline stages support transforming and
filtering log lines.</p>
</td>
</tr>
<tr>
<td>
<code>relabelings</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.RelabelConfig">
[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.RelabelConfig
</a>
</em>
</td>
<td>
<p>RelabelConfigs to apply to logs before delivering.
Grafana Agent Operator automatically adds relabelings for a few standard
Kubernetes fields and replaces original scrape job name with
__tmp_logs_job_name.</p>
<p>More info: <a href="https://grafana.com/docs/loki/latest/clients/promtail/configuration/#relabel_configs">https://grafana.com/docs/loki/latest/clients/promtail/configuration/#relabel_configs</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.QueueConfig">QueueConfig
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>QueueConfig allows the tuning of remote_write queue_config parameters.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>capacity</code><br/>
<em>
int
</em>
</td>
<td>
<p>Capacity is the number of samples to buffer per shard before we start dropping them.</p>
</td>
</tr>
<tr>
<td>
<code>minShards</code><br/>
<em>
int
</em>
</td>
<td>
<p>MinShards is the minimum number of shards, i.e. amount of concurrency.</p>
</td>
</tr>
<tr>
<td>
<code>maxShards</code><br/>
<em>
int
</em>
</td>
<td>
<p>MaxShards is the maximum number of shards, i.e. amount of concurrency.</p>
</td>
</tr>
<tr>
<td>
<code>maxSamplesPerSend</code><br/>
<em>
int
</em>
</td>
<td>
<p>MaxSamplesPerSend is the maximum number of samples per send.</p>
</td>
</tr>
<tr>
<td>
<code>batchSendDeadline</code><br/>
<em>
string
</em>
</td>
<td>
<p>BatchSendDeadline is the maximum time a sample will wait in buffer.</p>
</td>
</tr>
<tr>
<td>
<code>maxRetries</code><br/>
<em>
int
</em>
</td>
<td>
<p>MaxRetries is the maximum number of times to retry a batch on recoverable errors.</p>
</td>
</tr>
<tr>
<td>
<code>minBackoff</code><br/>
<em>
string
</em>
</td>
<td>
<p>MinBackoff is the initial retry delay. Gets doubled for every retry.</p>
</td>
</tr>
<tr>
<td>
<code>maxBackoff</code><br/>
<em>
string
</em>
</td>
<td>
<p>MaxBackoff is the maximum retry delay.</p>
</td>
</tr>
<tr>
<td>
<code>retryOnRateLimit</code><br/>
<em>
bool
</em>
</td>
<td>
<p>RetryOnRateLimit retries requests when encountering rate limits.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.RegexStageSpec">RegexStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>RegexStageSpec is a parsing stage that parses a log line using a regular
expression. Named capture groups in the regex allows for adding data into
the extracted map.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from extracted data to parse. If empty, defaults to using the log
message.</p>
</td>
</tr>
<tr>
<td>
<code>expression</code><br/>
<em>
string
</em>
</td>
<td>
<p>RE2 regular expression. Each capture group MUST be named. Required.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.RemoteWriteSpec">RemoteWriteSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.MetricsInstanceSpec">MetricsInstanceSpec</a>, <a href="#monitoring.grafana.com/v1alpha1.MetricsSubsystemSpec">MetricsSubsystemSpec</a>)
</p>
<div>
<p>RemoteWriteSpec defines the remote_write configuration for Prometheus.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the remote_write queue. Must be unique if specified. The name is
used in metrics and logging in order to differentiate queues.</p>
</td>
</tr>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<p>URL of the endpoint to send samples to.</p>
</td>
</tr>
<tr>
<td>
<code>remoteTimeout</code><br/>
<em>
string
</em>
</td>
<td>
<p>RemoteTimeout is the timeout for requests to the remote_write endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>headers</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Headers is a set of custom HTTP headers to be sent along with each
remote_write request. Be aware that any headers set by Grafana Agent
itself can&rsquo;t be overwritten.</p>
</td>
</tr>
<tr>
<td>
<code>writeRelabelConfigs</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.RelabelConfig">
[]github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.RelabelConfig
</a>
</em>
</td>
<td>
<p>WriteRelabelConfigs holds relabel_configs to relabel samples before they are
sent to the remote_write endpoint.</p>
</td>
</tr>
<tr>
<td>
<code>basicAuth</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.BasicAuth">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.BasicAuth
</a>
</em>
</td>
<td>
<p>BasicAuth for the URL.</p>
</td>
</tr>
<tr>
<td>
<code>bearerToken</code><br/>
<em>
string
</em>
</td>
<td>
<p>BearerToken used for remote_write.</p>
</td>
</tr>
<tr>
<td>
<code>bearerTokenFile</code><br/>
<em>
string
</em>
</td>
<td>
<p>BearerTokenFile used to read bearer token.</p>
</td>
</tr>
<tr>
<td>
<code>sigv4</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.SigV4Config">
SigV4Config
</a>
</em>
</td>
<td>
<p>SigV4 configures SigV4-based authentication to the remote_write endpoint.
Will be used if SigV4 is defined, even with an empty object.</p>
</td>
</tr>
<tr>
<td>
<code>tlsConfig</code><br/>
<em>
<a href="https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.TLSConfig">
github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1.TLSConfig
</a>
</em>
</td>
<td>
<p>TLSConfig to use for remote_write.</p>
</td>
</tr>
<tr>
<td>
<code>proxyUrl</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProxyURL to proxy requests through. Optional.</p>
</td>
</tr>
<tr>
<td>
<code>queueConfig</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.QueueConfig">
QueueConfig
</a>
</em>
</td>
<td>
<p>QueueConfig allows tuning of the remote_write queue parameters.</p>
</td>
</tr>
<tr>
<td>
<code>metadataConfig</code><br/>
<em>
<a href="#monitoring.grafana.com/v1alpha1.MetadataConfig">
MetadataConfig
</a>
</em>
</td>
<td>
<p>MetadataConfig configures the sending of series metadata to remote storage.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.ReplaceStageSpec">ReplaceStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>ReplaceStageSpec is a parsing stage that parses a log line using a regular
expression and replaces the log line. Named capture groups in the regex
allows for adding data into the extracted map.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from extracted data to parse. If empty, defaults to using the log
message.</p>
</td>
</tr>
<tr>
<td>
<code>expression</code><br/>
<em>
string
</em>
</td>
<td>
<p>RE2 regular expression. Each capture group MUST be named. Required.</p>
</td>
</tr>
<tr>
<td>
<code>replace</code><br/>
<em>
string
</em>
</td>
<td>
<p>Value to replace the captured group with.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.SigV4Config">SigV4Config
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.RemoteWriteSpec">RemoteWriteSpec</a>)
</p>
<div>
<p>SigV4Config specifies configuration to perform SigV4 authentication.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region of the AWS endpoint. If blank, the region from the default
credentials chain is used.</p>
</td>
</tr>
<tr>
<td>
<code>accessKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>AccessKey holds the secret of the AWS API access key to use for signing.
If not provided, The environment variable AWS_ACCESS_KEY_ID is used.</p>
</td>
</tr>
<tr>
<td>
<code>secretKey</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#secretkeyselector-v1-core">
Kubernetes core/v1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>SecretKey of the AWS API to use for signing. If blank, the environment
variable AWS_SECRET_ACCESS_KEY is used.</p>
</td>
</tr>
<tr>
<td>
<code>profile</code><br/>
<em>
string
</em>
</td>
<td>
<p>Profile is the named AWS profile to use for authentication.</p>
</td>
</tr>
<tr>
<td>
<code>roleARN</code><br/>
<em>
string
</em>
</td>
<td>
<p>RoleARN is the AWS Role ARN to use for authentication, as an alternative
for using the AWS API keys.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.TemplateStageSpec">TemplateStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>TemplateStageSpec is a transform stage that manipulates the values in the
extracted map using Go&rsquo;s template syntax.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from extracted data to parse. Required. If empty, defaults to using
the log message.</p>
</td>
</tr>
<tr>
<td>
<code>template</code><br/>
<em>
string
</em>
</td>
<td>
<p>Go template string to use. Required. In addition to normal template
functions, ToLower, ToUpper, Replace, Trim, TrimLeft, TrimRight,
TrimPrefix, and TrimSpace are also available.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.TenantStageSpec">TenantStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>TenantStageSpec is an action stage that sets the tenant ID for the log entry
picking it from a field in the extracted data map.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>label</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from labels whose value should be set as tenant ID. Mutually exclusive with
source and value.</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from extracted data to use as the tenant ID. Mutually exclusive with
label and value.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<p>Value to use for the template ID. Useful when this stage is used within a
conditional pipeline such as match. Mutually exclusive with label and source.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="monitoring.grafana.com/v1alpha1.TimestampStageSpec">TimestampStageSpec
</h3>
<p>
(<em>Appears on: </em><a href="#monitoring.grafana.com/v1alpha1.PipelineStageSpec">PipelineStageSpec</a>)
</p>
<div>
<p>TimestampStageSpec is an action stage that can change the timestamp of a log
line before it is sent to Loki.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>source</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name from extracted data to use as the timestamp. Required.</p>
</td>
</tr>
<tr>
<td>
<code>format</code><br/>
<em>
string
</em>
</td>
<td>
<p>Determines format of the time string. Required. Can be one of:
ANSIC, UnixDate, RubyDate, RFC822, RFC822Z, RFC850, RFC1123, RFC1123Z,
RFC3339, RFC3339Nano, Unix, UnixMs, UnixUs, UnixNs.</p>
</td>
</tr>
<tr>
<td>
<code>fallbackFormats</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Fallback formats to try if format fails.</p>
</td>
</tr>
<tr>
<td>
<code>location</code><br/>
<em>
string
</em>
</td>
<td>
<p>IANA Timezone Database string.</p>
</td>
</tr>
<tr>
<td>
<code>actionOnFailure</code><br/>
<em>
string
</em>
</td>
<td>
<p>Action to take when the timestamp can&rsquo;t be extracted or parsed.
Can be skip or fudge. Defaults to fudge.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>022b9c17</code>.
</em></p>
