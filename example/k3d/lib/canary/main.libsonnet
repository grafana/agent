local config = import 'config.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

// backwards compatibility with ksonnet
local envVar = if std.objectHasAll(k.core.v1, 'envVar') then k.core.v1.envVar else k.core.v1.container.envType;

local container = k.core.v1.container;
{
  new(namespace=''):: {
    local this = self,

    _images+:: {
      loki_canary: 'grafana/loki-canary:2.5.0',
    },

    loki_canary_args:: {
      labelvalue: '$(POD_NAME)',
      addr: 'loki',
      tls: false,
      port: 80,
      labelname: 'pod',
      interval: '500ms',
      size: 1024,
      wait: '1m',
      'metric-test-interval': '30m',
      'metric-test-range': '6h',
      user: 104334,
      pass: 'no-read-key',
    },

    container::
      container.new('loki-canary', this._images.loki_canary) +
      k.util.resourcesRequests('10m', '30Mi') +
      container.withPorts(k.core.v1.containerPort.new(name='http-metrics', port=80)) +
      container.withArgsMixin(k.util.mapToFlags(this.loki_canary_args)) +
      container.withEnv([
        envVar.fromFieldPath('HOSTNAME', 'spec.nodeName'),
        envVar.fromFieldPath('POD_NAME', 'metadata.name'),
      ]),
    
      local deployment = k.apps.v1.deployment,
      local service = k.core.v1.service,
    
      deployment: deployment.new('canary', 1, [this.container]) +
          	                  deployment.mixin.metadata.withNamespace(namespace),
      service:
          k.util.serviceFor(this.deployment) +
          service.mixin.metadata.withNamespace(namespace),
  }
}
