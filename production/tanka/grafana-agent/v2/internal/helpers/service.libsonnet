local k = import 'ksonnet-util/kausal.libsonnet';
local svc = k.core.v1.service;

{
  service(config={}):: {
    local this = self,
    local _config = this._config,

    controller_service:
      k.util.serviceFor(this.controller) +
      svc.mixin.metadata.withNamespace(_config.namespace),
  },
}
