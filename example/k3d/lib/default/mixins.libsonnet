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
agent_mixin {
  _config+: {
    // We run a single-node cortex so replace the job names to all
    // be the monolith.
    job_names+: {
      ingester: 'cortex',
      distributor: 'cortex',
      querier: 'cortex',
      query_frontend: 'cortex',
      table_manager: 'cortex',
      store_gateway: 'cortex',
      gateway: 'cortex',
    },
  },
}
