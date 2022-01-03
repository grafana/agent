local marshal = import 'ext/marshal.libsonnet';
local optionals = import 'ext/optionals.libsonnet';

// Returns the YAML config for the integration instance.
//
// @param {MetricsIntegrationInstanceSpec} spec.
local instance_config(spec) =
  local data = marshal.fromYAML(spec.Config);
  if data != null then data else {};

// Generates an integration instance.
//
// @param {MetricsIntegrationInstance} instance
function(instance) {
  local spec = instance.Spec,

  [key]: instance_config(spec) {
    // Normally integrations are disabled by default. However, the presence of
    // an IntegrationInstance implies that the user wishes to run it. We flip
    // the default here.
    //
    // There are some circumstances where users may wish to still disable the
    // integration, so we allow for an explicit override. Future changes may
    // remove this capability.
    enabled: if 'enabled' in super && super.enabled != null then super.enabled else true,

    // Force the integration to never scrape itself.
    // TODO: In the future, we may wish to remove this override.
    scrape_integration: false,
  }
  for key in [instance.Spec.Name]
}
