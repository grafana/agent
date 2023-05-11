local utils = import './internal/utils.jsonnet';

{
  // attr returns the field name that should be used for River attributes.
  attr(name):: name,

  // block returns the field name that should be used for River blocks.
  block(name, label='', index=0)::
    if label == ''
    then ('block %d %s' % [index, name])
    else ('block %d %s %s' % [index, name, label]),

  // expr returns an object which represents a literal River expression.
  expr(lit):: {
    // We need to use a special marker field to indicate that this object is an
    // expression, otherwise manifest.jsonnet would treat it as an object
    // literal.
    //
    // This field *must* be public. See utils.exprMarker for more information.
    [utils.exprMarker]: lit,
  },
}
