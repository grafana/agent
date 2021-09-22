{
  injectUtils(dashboard):: dashboard {
    tags: ['grafana-agent-mixin'],
    refresh: '30s',
    addMultiTemplateWithAll(name, metric_name, label_name, all='.*', hide=0):: self {
      templating+: {
        list+: [{
          allValue: all,
          current: {
            selected: true,
            text: 'All',
            value: '$__all',
          },
          datasource: '$datasource',
          hide: hide,
          includeAll: true,
          label: name,
          multi: true,
          name: name,
          options: [],
          query: 'label_values(%s, %s)' % [metric_name, label_name],
          refresh: 1,
          regex: '',
          sort: 2,
          tagValuesQuery: '',
          tags: [],
          tagsQuery: '',
          type: 'query',
          useTags: false,
        }],
      },
    },
  },
}
