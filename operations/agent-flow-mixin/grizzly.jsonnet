// grizzly.jsonnet allows you to test this mixin using Grizzly.
//
// To test, first set the GRAFANA_URL environment variable to the URL of a
// Grafana instance to deploy the mixin (i.e., "http://localhost:3000").
//
// Then, run `grr watch . ./grizzly.jsonnet` from this directory to watch the
// mixin and continually deploy all dashboards.
//
// Only dashboards get deployed; not alerts or recording rules.

local mixin = import './mixin.libsonnet';

{
  folder: {
    apiVersion: 'grizzly.grafana.com/v1alpha1',
    kind: 'DashboardFolder',
    metadata: {
      name: 'grafana-agent-flow',
    },
    spec: {
      title: mixin.grafanaDashboardFolder,
    },
  },

  dashboards: {
    [file]: {
      apiVersion: 'grizzly.grafana.com/v1alpha1',
      kind: 'Dashboard',
      metadata: {
        folder: $.folder.metadata.name,
        name: std.split(file, '.')[0],
      },
      spec: mixin.grafanaDashboards[file],
    }
    for file in std.objectFields(mixin.grafanaDashboards)
  },


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
