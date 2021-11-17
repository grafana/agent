local grafana_mixins = import 'default/mixins.libsonnet';
local datasource = import 'grafana/datasource.libsonnet';
local grafana = import 'grafana/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';
local kube_state_metrics = import 'kube-state-metrics/main.libsonnet';
local node_exporter = import 'node-exporter/main.libsonnet';
local prometheus = import 'prometheus/main.libsonnet';

local namespace = k.core.v1.namespace;
local ingress = k.networking.v1.ingress;
local rule = k.networking.v1.ingressRule;
local path = k.networking.v1.httpIngressPath;

local prometheus_monitoring = import './prometheus_monitoring.libsonnet';

{
  ns: namespace.new('monitoring'),

  grafana:
    grafana.new(namespace='monitoring') +
    grafana.withDashboards(grafana_mixins.grafanaDashboards) +
    grafana.withDataSources([
      datasource.new('Prometheus', 'http://prometheus.monitoring.svc.cluster.local:9090', default='true'),
      datasource.new('Cortex', 'http://cortex.smoke.svc.cluster.local/api/prom'),
    ]),

  prometheus:
    prometheus.new(namespace='monitoring') +
    prometheus.withConfigMixin(prometheus_monitoring.config) +
    prometheus.withRulesMixin(prometheus_monitoring.rules),

  node_exporter: node_exporter.new(namespace='monitoring'),
  kube_state_metrics: kube_state_metrics.new(namespace='monitoring'),

  ingresses: {
    prometheus:
      ingress.new('prometheus') +
      ingress.mixin.metadata.withNamespace('monitoring') +
      ingress.mixin.spec.withRules([
        rule.withHost('prometheus.k3d.localhost') +
        rule.http.withPaths([
          path.withPath('/') +
          path.withPathType('Prefix') +
          path.backend.service.withName('prometheus') +
          path.backend.service.port.withNumber(9090),
        ]),
      ]),

    grafana:
      ingress.new('grafana') +
      ingress.mixin.metadata.withNamespace('monitoring') +
      ingress.mixin.spec.withRules([
        rule.withHost('grafana.k3d.localhost') +
        rule.http.withPaths([
          path.withPath('/') +
          path.withPathType('Prefix') +
          path.backend.service.withName('grafana') +
          path.backend.service.port.withNumber(80),
        ]),
      ]),
  },
}
