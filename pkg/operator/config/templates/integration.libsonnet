local marshal = import 'ext/marshal.libsonnet';
local optionals = import 'ext/optionals.libsonnet';
local k8s = import 'utils/k8s.libsonnet';

// Returns the YAML config for the integration.
//
// @param {MetricsIntegrationSpec} spec
local instance_config(spec) =
  local raw = spec.Config.Raw;
  if raw == null || std.length(raw) == 0 then {}
  else (
    local data = marshal.fromRawJSON(spec.Config.Raw);
    if data != null then data else {}
  );

// Generates the individual config for the specified integration.
//
// @param {GrafanaAgent} agent
// @param {MetricsIntegration} integration
function(agent, integration) instance_config(integration.Spec) {
  // Force settings so the integration never scrapes itself. In the future we
  // may want to remove this restriction.
  autoscrape: { enable: false },

  extra_labels: {
    __meta_agentoperator_grafanaagent_name: agent.ObjectMeta.Name,
    __meta_agentoperator_grafanaagent_namespace: agent.ObjectMeta.Namespace,
    __meta_agentoperator_integration_type: integration.Spec.Type,
    __meta_agentoperator_integration_cr_name: integration.ObjectMeta.Name,
    __meta_agentoperator_integration_cr_namespace: integration.ObjectMeta.Namespace,
  } + std.foldl(
    function(acc, key) acc {
      local labels = integration.ObjectMeta.Labels,
      ['__meta_agentoperator_integration_cr_label_' + k8s.sanitize(key)]: labels[key],
      ['__meta_agentoperator_integration_cr_labelpresent_' + k8s.sanitize(key)]: 'true',
    },
    std.objectFields(if integration.ObjectMeta.Labels != null then integration.ObjectMeta.Labels else {}),
    {},
  ) + (
    if 'extra_labels' in super then super.extra_labels else {}
  ),
}
