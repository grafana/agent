function(replicas=1) {
  local this = self,
  local _config = this._config,
  local name = _config.name,
  local namespace = _config.namespace,

  local k = (import 'ksonnet-util/kausal.libsonnet') { _config+:: this._config },
  local deployment = k.apps.v1.deployment,

  controller:
    deployment.new(name, replicas, [this.container]) +
    deployment.mixin.metadata.withNamespace(namespace) +
    deployment.mixin.spec.template.spec.withServiceAccountName(name) +
    (
      if _config.config_hash
      then deployment.mixin.spec.template.metadata.withAnnotationsMixin({
        config_hash: std.md5(std.toString(_config.agent_config)),
      })
      else {}
    ) +
    k.util.configVolumeMount(name, '/etc/agent'),
}
