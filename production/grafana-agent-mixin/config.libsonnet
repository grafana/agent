{
  local makeGroupBy(groups) = std.join(', ', groups),

  _config+:: {
    namespace: '.*',

    // Groups labels to uniquely identify and group by clusters
    cluster_selectors: ['cluster', 'namespace'],

    // Each group-by label list is `, `-separated and unique identifies
    group_by_cluster: makeGroupBy($._config.cluster_selectors),
  },
}
