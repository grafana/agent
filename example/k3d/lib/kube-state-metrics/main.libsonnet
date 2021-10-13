local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local service = k.core.v1.service;
local serviceAccount = k.core.v1.serviceAccount;
local policyRule = k.rbac.v1.policyRule;

{
  new(namespace=''):: {
    local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: { namespace: namespace } },

    container::
      container.new('kube-state-metrics', 'k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.1.0') +
      container.withPorts([
        containerPort.newNamed(name='http-metrics', containerPort=8080),
        containerPort.newNamed(name='self-metrics', containerPort=8081),
      ]) +
      container.withArgs([
        '--port=8080',
        '--telemetry-host=0.0.0.0',
        '--telemetry-port=8081',
      ]),

    rbac:
      k.util.rbac('kube-state-metrics', [
        policyRule.withApiGroups(['']) +
        policyRule.withResources([
          'configmaps',
          'secrets',
          'nodes',
          'pods',
          'services',
          'resourcequotas',
          'replicationcontrollers',
          'limitranges',
          'persistentvolumeclaims',
          'persistentvolumes',
          'namespaces',
          'endpoints',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['extensions']) +
        policyRule.withResources([
          'daemonsets',
          'deployments',
          'replicasets',
          'ingresses',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['apps']) +
        policyRule.withResources([
          'daemonsets',
          'deployments',
          'replicasets',
          'statefulsets',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['batch']) +
        policyRule.withResources([
          'cronjobs',
          'jobs',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['autoscaling']) +
        policyRule.withResources([
          'horizontalpodautoscalers',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['authorization.k8s.io']) +
        policyRule.withResources(['subjectaccessreviews']) +
        policyRule.withVerbs(['create']),

        policyRule.withApiGroups(['ingresses']) +
        policyRule.withResources(['ingress']) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['policy']) +
        policyRule.withResources(['poddisruptionbudgets']) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['certificates.k8s.io']) +
        policyRule.withResources(['certificatesigningrequests']) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['storage.k8s.io']) +
        policyRule.withResources([
          'storageclasses',
          'volumeattachments',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['admissionregistration.k8s.io']) +
        policyRule.withResources([
          'mutatingwebhookconfigurations',
          'validatingwebhookconfigurations',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['networking.k8s.io']) +
        policyRule.withResources([
          'networkpolicies',
          'ingresses',
        ]) +
        policyRule.withVerbs(['list', 'watch']),

        policyRule.withApiGroups(['coordination.k8s.io']) +
        policyRule.withResources(['leases']) +
        policyRule.withVerbs(['list', 'watch']),
      ]) {
        service_account+:
          serviceAccount.mixin.metadata.withNamespace(namespace),
      },

    deployment:
      deployment.new('kube-state-metrics', 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace) +
      deployment.mixin.spec.template.metadata.withAnnotationsMixin({ 'prometheus.io.scrape': 'false' }) +
      deployment.mixin.spec.template.spec.withServiceAccount('kube-state-metrics') +
      deployment.mixin.spec.template.spec.securityContext.withRunAsUser(65534) +
      deployment.mixin.spec.template.spec.securityContext.withRunAsGroup(65534) +
      deployment.mixin.spec.template.spec.securityContext.withFsGroup(0) +
      k.util.podPriority('critical'),

    service:
      k.util.serviceFor(self.deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },
}
