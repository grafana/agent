function() {
  local this = self,
  local _config = this._config,
  local name = _config.name,
  local namespace = _config.namespace,

  local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: this._config },
  local daemonSet = k.apps.v1.daemonSet,

  controller:
    daemonSet.new(name, [this.container]) +
    daemonSet.mixin.metadata.withNamespace(namespace) +
    daemonSet.mixin.spec.template.spec.withServiceAccountName(name) +
    (
      if _config.config_hash
      then daemonSet.mixin.spec.template.metadata.withAnnotationsMixin({
        config_hash: std.md5(std.toString(_config.agent_config)),
      })
      else {}
    ) +
    k.util.configVolumeMount(name, '/etc/agent'),
}
