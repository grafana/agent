{
  // group groups the set of input integrations based on the integration name.
  // The result is multiple sets of integrations where each set holds
  // integrations of the same name.
  //
  // @param {MetricsIntegration[]} set
  // @returns {MetricsIntegration[][]}
  group(set)::
    // Group into an object by using the integration name as the key.
    local map = std.foldl(
      function(acc, element) acc {
        [element.Spec.Name]: (
          local key = element.Spec.Name;
          if std.objectHas(acc, key) then (
            assert $.integrationsMatch(acc[key][0], element) : 'integrations do not match';
            acc[key] + [element]
          ) else [element]
        ),
      }, set, {},
    );
    // Then flatten our object into an array.
    std.foldl(
      function(acc, key) acc + [map[key]],
      std.objectFields(map),
      [],
    ),

  // Returns true if a and b have the same name and type.
  //
  // @param {MetricsIntegration} a
  // @param {MetricsIntegration} b
  // @returns {Boolean}
  integrationsMatch(a, b)::
    a.Spec.Type == b.Spec.Type &&
    a.Spec.Name == b.Spec.Name,

  // groupName returns the name of a group. It is assumed that every element in
  // group is of the same integration.
  //
  // @param {MetricsIntegration[]} group
  // @returns {String}
  groupName(group)::
    assert std.length(group) > 0 : 'unexpected empty group of integrations';
    if group[0].Spec.Type == 'normal' then group[0].Spec.Name + '_configs'
    else group[0].Spec.Name,
}
