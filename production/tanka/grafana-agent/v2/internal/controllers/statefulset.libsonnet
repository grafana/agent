function(replicas=1, volumeClaims=[]) {
  local this = self,
  local _config = this._config,
  local name = _config.name,
  local namespace = _config.namespace,

  local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: this._config },
  local statefulSet = k.apps.v1.statefulSet,

  controller:
    statefulSet.new(name, replicas, [this.container], volumeClaims) +
    statefulSet.mixin.metadata.withNamespace(namespace) +
    statefulSet.mixin.spec.withServiceName(name) +
    statefulSet.mixin.spec.template.spec.withServiceAccountName(name) +
    (
      if _config.config_hash
      then statefulSet.mixin.spec.template.metadata.withAnnotationsMixin({
        config_hash: std.md5(std.toString(_config.agent_config)),
      })
      else {}
    ) +
    k.util.configVolumeMount(name, '/etc/agent'),
}
