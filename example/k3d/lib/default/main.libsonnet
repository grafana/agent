local cortex = import 'cortex/main.libsonnet';
local datasource = import 'grafana/datasource.libsonnet';
local grafana = import 'grafana/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';
local loki = import 'loki/main.libsonnet';
local metrics = import 'metrics-server/main.libsonnet';

local mixins = import './mixins.libsonnet';

{
  new(namespace=''):: {
    ns: k.core.v1.namespace.new(namespace),

    grafana:
      grafana.new(namespace=namespace) +
      grafana.withDashboards(mixins.grafanaDashboards) +
      grafana.withDataSources([
        datasource.new('Cortex', 'http://cortex.default.svc.cluster.local/api/prom'),
        datasource.new('Loki', 'http://loki.default.svc.cluster.local', type='loki'),
      ]),

    loki: loki.new(namespace),

    cortex: cortex.new(namespace),
  },
}
