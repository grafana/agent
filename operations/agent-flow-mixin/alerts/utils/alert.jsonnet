// alert.jsonnet defines utilities to create alerts.

{
  new(rules):: {
    rules: rules,
  },

  newRule(name='', expr='', message='', forT=''):: std.prune({
    alert: name,
    expr: expr,
    annotations: {
      message: message,
    },
    'for': forT,
  }),
}
