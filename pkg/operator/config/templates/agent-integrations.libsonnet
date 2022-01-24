// agent-integrations.libsonnet is the entrypoint for rendering a Grafana Agent
// config file for integrations based on the Operator custom resources.
//
// When writing an object, any field will null will be removed from the final
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

local new_integration = import './integration.libsonnet';
local integrations = import 'utils/integrations.libsonnet';

// @param {Hierarchy} ctx
function(ctx) marshal.YAML(optionals.trim({
  local spec = ctx.Agent.Spec,
  local prometheus = spec.Metrics,
  local namespace = ctx.Agent.ObjectMeta.Namespace,

  server: {
    http_listen_port: 8080,
    log_level: optionals.string(spec.LogLevel),
    log_format: optionals.string(spec.LogFormat),
  },

  integrations: {
    metrics: {
      autoscrape: {
        enable: false,
      },
    },
  } + {
    [integrations.groupName(group)]: (
      if group[0].Spec.Type == 'normal'
      then [new_integration(ctx.Agent, integration) for integration in group]
      else (
        assert std.length(group) == 1 : 'non-normal integration can only have 1 instance';
        new_integration(ctx.Agent, group[0])
      )
    )
    for group in integrations.group(ctx.Integrations)
  },
}))
