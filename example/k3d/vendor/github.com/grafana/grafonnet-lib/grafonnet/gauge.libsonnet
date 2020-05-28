{
  /**
   * @name gauge.new
   */
  new(
    title,
    datasource=null,
    calc='mean',
    time_from=null,
    span=null,
    description='',
    height=null,
    transparent=null,
  )::
    {
      [if description != '' then 'description']: description,
      [if height != null then 'height']: height,
      [if transparent != null then 'transparent']: transparent,
      [if time_from != null then 'timeFrom']: time_from,
      [if span != null then 'span']: span,
      title: title,
      type: 'gauge',
      datasource: datasource,
      options: {
        fieldOptions: {
          calcs: [
            calc,
          ],
        },
      },
      _nextTarget:: 0,
      addTarget(target):: self {
        local nextTarget = super._nextTarget,
        _nextTarget: nextTarget + 1,
        targets+: [target { refId: std.char(std.codepoint('A') + nextTarget) }],
      },
    },

}
