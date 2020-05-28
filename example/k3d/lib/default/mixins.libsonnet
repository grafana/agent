local cortex_mixin = import 'cortex-mixin/mixin.libsonnet';
local agent_debugging_mixin = import 'grafana-agent-mixin/debugging.libsonnet';
local agent_mixin = import 'grafana-agent-mixin/mixin.libsonnet';

// TODO(rfratto): bit of a hack here to be compatible with the "old" Jsonnet
// writing style.
local fix = {
  dashboards+:: {},
  grafana_dashboards+:: {},
  grafanaDashboards+:: $.dashboards + $.grafana_dashboards,
};

fix +
cortex_mixin +
agent_debugging_mixin +
agent_mixin
