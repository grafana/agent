{
  /**
   * Returns a new table panel that can be added in a row.
   * It requires the table panel plugin in grafana, which is built-in.
   *
   * @param title The title of the graph panel.
   * @param span Width of the panel
   * @param height Height of the panel
   * @param description Description of the panel
   * @param datasource Datasource
   * @param min_span Min span
   * @param styles Styles for the panel
   * @param columns Columns for the panel
   * @param sort Sorting instruction for the panel
   * @param transform allow table manipulation to present data as desired
   * @param transparent Boolean (default: false) If set to true the panel will be transparent
   * @return A json that represents a table panel
   */
  new(
    title,
    description=null,
    span=null,
    min_span=null,
    height=null,
    datasource=null,
    styles=[],
    transform=null,
    transparent=false,
    columns=[],
    sort=null,
    time_from=null,
    time_shift=null,
  ):: {
    type: 'table',
    title: title,
    [if span != null then 'span']: span,
    [if min_span != null then 'minSpan']: min_span,
    [if height != null then 'height']: height,
    datasource: datasource,
    targets: [
    ],
    styles: styles,
    columns: columns,
    timeFrom: time_from,
    timeShift: time_shift,
    [if sort != null then 'sort']: sort,
    [if description != null then 'description']: description,
    [if transform != null then 'transform']: transform,
    [if transparent == true then 'transparent']: transparent,

    _nextTarget:: 0,
    addTarget(target):: self + self.addTargets([target]),
    addTargets(newtargets)::
      self {
        local n = std.foldl(function(numOfTargets, p)
          (if 'targets' in p then
             numOfTargets + 1 + std.length(p.targets)
           else
             numOfTargets + 1), newtargets, 0),
        local nextTarget = super._nextTarget,
        local _targets = std.makeArray(
          std.length(newtargets), function(i)
            newtargets[i] {
              refId: std.char(std.codepoint('A') + nextTarget + (
                if i == 0 then
                  0
                else
                  if 'targets' in _targets[i - 1] then
                    (std.codepoint(_targets[i - 1].refId) - nextTarget) + 1 + std.length(_targets[i - 1].targets)
                  else
                    (std.codepoint(_targets[i - 1].refId) - nextTarget) + 1
              )),
              [if 'targets' in newtargets[i] then 'targets']: std.makeArray(
                std.length(newtargets[i].targets), function(j)
                  newtargets[i].targets[j] {
                    refId: std.char(std.codepoint('A') + 1 + j +
                                    nextTarget + (
                      if i == 0 then
                        0
                      else
                        if 'targets' in _targets[i - 1] then
                          (std.codepoint(_targets[i - 1].refId) - nextTarget) + 1 + std.length(_targets[i - 1].targets)
                        else
                          (std.codepoint(_targets[i - 1].refId) - nextTarget) + 1
                    )),
                  }
              ),
            }
        ),

        _nextTarget: nextTarget + n,
        targets+::: _targets,
      },
    addColumn(field, style):: self {
      local style_ = style { pattern: field },
      local column_ = { text: field, value: field },
      styles+: [style_],
      columns+: [column_],
    },
    hideColumn(field):: self {
      styles+: [{
        alias: field,
        pattern: field,
        type: 'hidden',
      }],
    },
  },
}
