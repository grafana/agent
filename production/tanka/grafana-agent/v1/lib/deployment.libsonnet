local agent = import '../internal/agent.libsonnet';

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

    config:: {
      server: {
        log_level: 'info',
        http_listen_port: 8080,
      },

      prometheus:
        if !has_prometheus_config then {}
        else this._prometheus_config {
          configs:
            if has_prometheus_instances
            then this._prometheus_instances
            else [],
        },
    },

    agent:
      agent.newAgent(name, namespace, self._images.agent, self.config, use_daemonset=true) +
      agent.withConfigHash(self._config_hash),
  },
}
