// Generates an external_label mapping. This includes the
// user-provided labels as well as the injected cluster and
// replica labels.
//
// @param {config.Deployment} ctx
function(ctx) (
  local meta = ctx.Agent.ObjectMeta;
  local prometheus = ctx.Agent.Spec.Prometheus;

  // Provide the cluster label first. Doing it this way allows the user to
  // override with a value they choose.
  (
    local clusterValue = '%s/%s' % [meta.Namespace, meta.Name];
    local clusterLabel = prometheus.PrometheusExternalLabelName;

    if clusterLabel == null then { cluster: clusterValue }
    else if clusterLabel != '' then { [clusterLabel]: clusterValue }
    else {}
  ) +

  // Then add in any user-configured labels.
  (
    if prometheus.ExternalLabels == null then {}
    else prometheus.ExternalLabels
  ) +

  // Finally, add the replica label. We don't want the user to overrwrite the
  // replica label since it can cause duplicate sample problems.
  (
    local replicaValue = 'replica-$(STATEFULSET_ORDINAL_NUMBER)';
    local replicaLabel = prometheus.ReplicaExternalLabelName;

    if replicaLabel == null then { __replica__: replicaValue }
    else if replicaLabel != '' then { [replicaLabel]: replicaValue }
    else {}
  )


)
