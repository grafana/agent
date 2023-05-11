local utils = import './internal/utils.jsonnet';

// parseField parses a field name from a Jsonnet object and determines if it's
// supposed to be a River attribute or block.
local parseField(name) = (
  local parts = std.split(name, ' ');
  local numParts = std.length(parts);
  if numParts == 1 then {
    type: 'attr',
    name: parts[0],
    index: 0,
    orig: name,
  }
  else if numParts > 1 && numParts <= 4 && parts[0] == 'block' then {
    type: 'block',
    index: std.parseInt(parts[1]),
    name: parts[2],
    label: if numParts == 4 then parts[3] else '',
    orig: name,
  } else (
    error 'invalid field name %s' % name
  )
);

// isRiverExpr returns true if value was constructed with river.expr().
local isRiverExpr(value) = std.isObject(value) && std.length(value) == 1 && utils.exprMarker in value;

// linePadding returns number of spaces to indent a line with given a specific
// indentation level.
local linePadding(indent) = std.repeat('  ', indent);

local manifester(indent=0) = {
  local padding = linePadding(indent),
  local this = $,

  // body manifests a River body to text, manifesting both attributes and
  // blocks.
  body(value): (
    // First we need to look at each public field of the value and parse it as
    // an attribute or block.
    local parsedFields = std.sort(
      std.map(function(field) parseField(field), std.objectFields(value)),
      function(field) field.index,
    );

    // Now we can accumulate all of the fields into a single string. Each field
    // will be separated by exactly one newline.
    std.foldl(function(acc, field) (
      // Determine the name to use for the field.
      local name = (
        if field.type == 'attr' then field.name
        else if field.type == 'block' && field.label == '' then field.name
        else '%s "%s"' % [field.name, field.label]
      );

      if field.type == 'attr' then (
        // Manifest the value to text.
        local attr_value = this.value(value[field.orig]);

        // Attributes are printed as <attribute_name> = <rendered value>
        acc + padding + ('%s = %s' % [name, attr_value]) + '\n'
      ) else if field.type == 'block' && std.isObject(value[field.orig]) then (
        local block_header = '%s {\n' % name;
        local block_body = manifester(indent + 1).body(value[field.orig]);
        // The block body ends in a newline, so the block trailer must start
        // with line padding.
        local block_trailer = padding + '}';

        acc + padding + block_header + block_body + block_trailer + '\n'
      ) else if field.type == 'block' && std.isArray(value[field.orig]) then (
        // List of blocks.

        local block_header = '%s {\n' % name;
        local block_trailer = padding + '}';

        std.foldl(function(acc, block) (
          local block_body = manifester(indent + 1).body(block);
          acc + padding + block_header + block_body + block_trailer + '\n'
        ), value[field.orig], acc)
      ) else (
        error 'invalid field type'  // This should never happen
      )
    ), parsedFields, '')
  ),

  // value manifests a River value to text.
  value(value): (
    if value == null then (
      'null'
    ) else if isRiverExpr(value) then (
      local lines = std.split(value[utils.exprMarker], '\n');

      // When injecting literals, each line after the first should have the
      // current padding appended to it.
      std.join('\n', std.mapWithIndex(function(index, line) (
        if index > 0 then padding + line else line
      ), lines))
    ) else if std.isString(value) then (
      '"%s"' % value
    ) else if std.isBoolean(value) then (
      std.toString(value)
    ) else if std.isNumber(value) then (
      std.toString(value)
    ) else if std.isArray(value) then (
      // To manifest an array, we convert all of the elements into text and
      // separate them by commas.

      // First pair the elements with their index and value so we know when we
      // need commas.
      local elements = std.mapWithIndex(function(index, elem) {
        index: index,
        value: elem,
      }, value);

      // Finally, construct the body to insert in between the square brackets.
      local body = std.foldl(function(acc, elem) (
        // We need to put a comma at the end if we're not the last element.
        local suffix =
          if elem.index + 1 < std.length(elements)
          then ', '
          else '';

        local text = this.value(elem.value);

        acc + text + suffix
      ), elements, '');

      '[%s]' % body
    ) else if std.isObject(value) then (
      // To manifest an object, we convert all of the fields into text and add
      // a comma and a newline after each. Unlike arrays, this always has a
      // trailing comma so we don't need to track the index.

      // Create an inner manifester with a higher indentation level for
      // printing the fields.
      local next = manifester(indent + 1);

      local body = std.foldl(function(acc, field) (
        // TODO(rfratto): detect whether it's necessary to wrap the field name
        // in quotes. This is only necessary if it's not a valid identifier.
        local text = '"%s" = %s,\n' % [field, next.value(value[field])];

        // Make sure the attribute itself is given the padding from the next
        // indentation level, not the current one.
        acc + linePadding(indent + 1) + text
      ), std.objectFields(value), '');

      // Last check at the end: if the body is empty, we can return an empty object.
      if body == '' then '{}'
      else (
        '{\n' + body + padding + '}'
      )
    )
    else error 'unsupported value type %s' % std.type(value)
  ),
};

{

  // manifestRiver returns a pretty-printed River file from the Jsonnet value.
  // value must be an object.
  manifestRiver(value):: (
    assert std.isObject(value) : 'manifestRiver must be called with object';
    manifester().body(value)
  ),

  manifestRiverValue(value):: (
    manifester().value(value)
  ),
}
