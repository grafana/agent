local mixin = import '../mixin.libsonnet';

{
  prometheus_rules: {
    [file]: {
      apiVersion: 'grizzly.grafana.com/v1alpha1',
      kind: 'PrometheusRuleGroup',
      metadata: {
        folder: $.folder.metadata.name,
        namespace: 'agent-flow',
        name: std.split(file, '.')[0],
      },
      spec: mixin.prometheusAlerts[file],
    }
    for file in std.objectFields(mixin.prometheusAlerts)
  },
}
