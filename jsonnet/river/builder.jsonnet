{
  // attr returns the field name that should be used for River attributes.
  attr(name):: name,

  // block returns the field name that should be used for River blocks.
  block(name, label='')::
    if label == ''
    then ('block %s' % name)
    else ('block %s %s' % [name, label]),

  // expr returns an object which reprents a literal River expression.
  expr(lit):: {
    _river_expr:: true,
    lit: lit,
  },
}
