local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;

{
  // withIntegrations controls the integrations component of the Agent.
  //
  // For the full list of options, refer to the configuration reference:
  // https://github.com/grafana/agent/blob/main/docs/configuration-reference.md#integrations_config
  withIntegrations(integrations):: {
    assert std.objectHasAll(self, '_mode') : |||
      withIntegrations must be merged with the result of calling new.
    |||,
    _integrations:: integrations,
  },

  integrationsMixin:: {
    container+::
      container.mixin.securityContext.withPrivileged(true) +
      container.mixin.securityContext.withRunAsUser(0),

    local controller = self._controller,
    agent+:
      // procfs, sysfs, rotfs
      k.util.hostVolumeMount('proc', '/proc', '/host/proc', readOnly=true) +
      k.util.hostVolumeMount('sys', '/sys', '/host/sys', readOnly=true) +
      k.util.hostVolumeMount('root', '/', '/host/root', readOnly=true) +

      controller.mixin.spec.template.spec.withHostPID(true) +
      controller.mixin.spec.template.spec.withHostNetwork(true) +
      controller.mixin.spec.template.spec.withDnsPolicy('ClusterFirstWithHostNet'),
  },
}
