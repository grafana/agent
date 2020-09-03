local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;

{
  // withIntegrations controls the integrations component of the Agent.
  //
  // For the full list of options, refer to the configuration reference:
  // https://github.com/grafana/agent/blob/master/docs/configuration-reference.md#integrations_config
  withIntegrations(integrations):: {
    assert std.objectHasAll(self, '_mode') : |||
      withLokiConfig must be merged with the result of calling new.
    |||,
    _integrations:: integrations,
  },

  // TODO(rfratto): only enable this when node_exporter is used.
  integrationsMixin:: {
    container+::
      container.mixin.securityContext.withPrivileged(true) +
      container.mixin.securityContext.withRunAsUser(0),

    local controller = self.agent._controller,
    agent+::
      // procfs, sysfs, rotfs
      k.util.hostVolumeMount('proc', '/proc', '/host/proc', readOnly=true) +
      k.util.hostVolumeMount('sys', '/sys', '/host/sys', readOnly=true) +
      k.util.hostVolumeMount('root', '/', '/host/root', readOnly=true) +

      controller.mixin.spec.template.spec.withHostPid(true) +
      controller.mixin.spec.template.spec.withHostNetwork(true),
  },
}
