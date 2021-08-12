local agent = import '../internal/agent.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local configMap = k.core.v1.configMap;
local service = k.core.v1.service;
local container = k.core.v1.container;

{
  // newDeployment creates a new single-replicaed Deployment of the
  // grafana-agent. By default, this deployment will do no collection. You must
  // merge the result of this function with the following:
  //
  // - withPrometheusConfig
  // - withPrometheusInstances
  // - optionally withRemoteWrite
  //
  // newDeployment does not support log collection.
  newDeployment(name='grafana-agent', namespace='default'):: {
    assert !std.objectHas(self, '_loki_config') : |||
      Log collection is not supported with newDeployment.
    |||,
    assert !std.objectHas(self, '_integrations') : |||
      Integrations are not supported with newDeployment.
    |||,

    local this = self,

    _mode:: 'deployment',
    _images:: $._images,
    _config_hash:: true,

    local has_prometheus_config = std.objectHasAll(self, '_prometheus_config'),
    local has_prometheus_instances = std.objectHasAll(self, '_prometheus_instances'),
    local has_tempo_config = std.objectHasAll(self, '_tempo_config'),
    local has_sampling_strategies = std.objectHasAll(self, '_tempo_sampling_strategies'),

    config:: {
      server: {
        log_level: 'info',
        http_listen_port: 8080,
      },
    } + (
      if has_prometheus_config
      then {
        prometheus:
          this._prometheus_config {
            configs:
              if has_prometheus_instances
              then this._prometheus_instances
              else [],
          },
      }
      else {}
    ) + (
      if has_tempo_config then {
        tempo: {
          configs: [this._tempo_config {
            name: 'default',
          }],
        },
      }
      else {}
    ),

    agent:
      agent.newAgent(name, namespace, self._images.agent, self.config, use_daemonset=false) +
      agent.withConfigHash(self._config_hash) + {
        // If sampling strategies were defined, we need to mount them as a JSON
        // file.
        config_map+:
          if has_sampling_strategies
          then configMap.withDataMixin({
            'strategies.json': std.toString(this._tempo_sampling_strategies),
          })
          else {},
        // If we're deploying for tracing, applications will want to write to
        // a service for load balancing span delivery.
        service:
          if has_tempo_config
          then k.util.serviceFor(self.agent) + service.mixin.metadata.withNamespace(namespace)
          else {},
      },
  },
}
