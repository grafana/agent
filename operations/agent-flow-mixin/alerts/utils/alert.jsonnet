// alert.jsonnet defines utilities to create alerts.

{
  new(rules):: {
    rules: rules,
  },

  newRule(name='', expr='', message=''):: std.prune({
    alert: name,
    expr: expr,
    annotations: {
      message: message,
    },
  }),

  withForTime(forTime=''):: { 'for': forTime },
}
