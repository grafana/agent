// dashboard.jsonnet defines utilities to create dashboards using the
// schemaVersion present in Grafana 9.

{
  new(name=''):: {
    title: name,
    timezone: 'utc',
    refresh: '10s',
    schemaVersion: 36,
    tags: ['grafana-agent-flow-mixin'],
    templating: {
      list: [{
        name: 'datasource',
        label: 'Data Source',
        type: 'datasource',
        query: 'prometheus',
        refresh: 1,
        sort: 2,
      }, {
        name: 'loki_datasource',
        label: 'Loki Data Source',
        type: 'datasource',
        query: 'loki',
        refresh: 1,
        sort: 2,
      }],
    },
    time: {
      from: 'now-1h',
      to: 'now',
    },
    timepicker: {
      refresh_intervals: [
        '5s',
        '10s',
        '30s',
        '1m',
        '5m',
        '15m',
        '30m',
        '1h',
        '2h',
        '1d',
      ],
      time_options: [
        '5m',
        '15m',
        '1h',
        '6h',
        '12h',
        '24h',
        '2d',
        '7d',
        '30d',
        '90d',
      ],
    },
  },

  withUID(uid):: { uid: uid },

  withTemplateVariablesMixin(vars):: {
    templating+: {
      list+: vars,
    },
  },

  newTemplateVariable(name, query):: {
    name: name,
    label: name,
    type: 'query',
    query: {
      query: query,
      refId: name,
    },
    datasource: '${datasource}',
    refresh: 2,
    sort: 2,
  },

  newLokiAnnotation(name, expression, color):: {
    name: name,
    datasource: '$loki_datasource',
    enable: true,
    expr: expression,
    iconColor: color,
    instant: false,
    titleFormat: '{{cluster}}/{{namespace}}',
  },

  newMultiTemplateVariable(name, query):: $.newTemplateVariable(name, query) {
    allValue: '.*',
    includeAll: true,
    multi: true,
  },

  withPanelsMixin(panels):: { panels+: panels },

  withAnnotations(annotations):: {
    annotations+: {
      list+: annotations,
    },
  },

  withDocsLink(url, desc):: {
    links+: [{
      title: 'Documentation',
      icon: 'doc',
      targetBlank: true,
      tooltip: desc,
      type: 'link',
      url: url,
    }],
  },

  withDashboardsLink():: {
    links+: [{
      title: 'Dashboards',
      type: 'dashboards',
      asDropdown: true,
      icon: 'external link',
      includeVars: true,
      keepTime: true,
      tags: ['grafana-agent-flow-mixin'],
      targetBlank: false,
    }],
  },
}
