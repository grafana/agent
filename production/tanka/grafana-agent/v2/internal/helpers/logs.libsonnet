local k = import 'ksonnet-util/kausal.libsonnet';
local container = k.core.v1.container;

{
  volumeMounts(config={}):: {
    // Disable journald mount by default
    local _config = {
      journald: false,
    } + config,

    controller+:
      // For reading docker containers. /var/log is used for the positions file
      // and shouldn't be set to readonly.
      k.util.hostVolumeMount('varlog', '/var/log', '/var/log') +
      k.util.hostVolumeMount('varlibdockercontainers', '/var/lib/docker/containers', '/var/lib/docker/containers', readOnly=true) +

      // For reading journald
      if _config.journald == false then {}
      else k.util.hostVolumeMount('etcmachineid', '/etc/machine-id', '/etc/machine-id', readOnly=true),
  },

  permissions(config={}):: {
    container+::
      container.mixin.securityContext.withPrivileged(true) +
      container.mixin.securityContext.withRunAsUser(0),
  },
}
