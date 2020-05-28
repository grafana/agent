{
  /**
   * @name prometheus.target
   */
  target(
    expr,
    format='time_series',
    intervalFactor=2,
    legendFormat='',
    datasource=null,
    interval=null,
    instant=null,
    hide=null,
  ):: {
    [if hide != null then 'hide']: hide,
    [if datasource != null then 'datasource']: datasource,
    expr: expr,
    format: format,
    intervalFactor: intervalFactor,
    legendFormat: legendFormat,
    [if interval != null then 'interval']: interval,
    [if instant != null then 'instant']: instant,
  },
}
