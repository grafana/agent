// agent-integration.libsonnet is the entrypoint for rendering a Grafana Agent
// config file for a single integration based on the Operator custom resources.
//
// When writing an object, any field with null will be removed from the final
// YAML. This is useful as we don't want to always translate unfilled values
// from the custom resources to a field in the YAML.
//
// A series of helper methods to convert default values into null (so they can
// be trimmed) are in ./ext/optionals.libsonnet.
//
// When writing a new function, please document the expected types of the
// arguments.

local marshal = import 'ext/marshal.libsonnet';
local optionals = import 'ext/optionals.libsonnet';

local new_integration_instance = import './integration.libsonnet';

// @param {config.Deployment} ctx
function(ctx) marshal.YAML(optionals.trim({
  local spec = ctx.Agent.Spec,
  local logs = spec.Logs,
  local namespace = ctx.Agent.ObjectMeta.Namespace,

  server: {
    http_listen_port: 8080,
    log_level: optionals.string(spec.LogLevel),
    log_format: optionals.string(spec.LogFormat),
  },

  integrations:
    if ctx.Integration != null then new_integration_instance(ctx.Integration),
}))
