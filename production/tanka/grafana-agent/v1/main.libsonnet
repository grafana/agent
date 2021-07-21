local agent = import './internal/agent.libsonnet';
local utils = import './internal/utils.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;
local configMap = k.core.v1.configMap;
local service = k.core.v1.service;

// Merge all of our libraries to create the final exposed library.
(import './lib/deployment.libsonnet') +
(import './lib/integrations.libsonnet') +
(import './lib/prometheus.libsonnet') +
(import './lib/scraping_service.libsonnet') +
(import './lib/loki.libsonnet') +
(import './lib/tempo.libsonnet') +
{
  _images:: {
    agent: 'grafana/agent:v0.16.1',
    agentctl: 'grafana/agentctl:v0.16.1',
  },

  // new creates a new DaemonSet deployment of the grafana-agent. By default,
  // the deployment will do no collection. You must merge the result of this
  // function with one or more of the following:
  //
  // - withPrometheusConfig, withPrometheusInstances (and optionally withRemoteWrite)
  // - withLokiConfig
  //
  // When using withPrometheusInstances, a [name]-etc deployment
  // with one replica will be created alongside the DaemonSet. This deployment
  // is responsible for handling scrape configs that will not work on the host
  // machine.
  //
  // For example, if a scrape_config scrapes the Kubernetes API, that must be
  // handled by the [name]-etc deployment as the Kubernetes API does not run
  // on any node in the cluster.
  //
  // scrapeInstanceKubernetes provides the default
  // PrometheusInstanceConfig Grafana Labs uses in production.
  new(name='grafana-agent', namespace='default'):: {
    local this = self,

    _mode:: 'daemonset',
    _images:: $._images,
    _config_hash:: true,

    local has_loki_config = std.objectHasAll(self, '_loki_config'),
    local has_tempo_config = std.objectHasAll(self, '_tempo_config'),
    local has_prometheus_config = std.objectHasAll(self, '_prometheus_config'),
    local has_prometheus_instances = std.objectHasAll(self, '_prometheus_instances'),
    local has_integrations = std.objectHasAll(self, '_integrations'),
    local has_sampling_strategies = std.objectHasAll(self, '_tempo_sampling_strategies'),

    local prometheus_instances =
      if has_prometheus_instances then this._prometheus_instances else [],
    local host_filter_instances = utils.transformInstances(prometheus_instances, true),
    local etc_instances = utils.transformInstances(prometheus_instances, false),

    config:: {
      server: {
        log_level: 'info',
        http_listen_port: 8080,
      },
    } + (
      if has_prometheus_config
      then { prometheus: this._prometheus_config { configs: host_filter_instances } }
      else {}
    ) + (
      if has_loki_config then {
        loki: {
          positions_directory: '/tmp/positions',
          configs: [this._loki_config {
            name: 'default',
          }],
        },
      } else {}
    ) + (
      if has_tempo_config then {
        tempo: {
          configs: [this._tempo_config {
            name: 'default',
          }],
        },
      }
      else {}
    ) + (
      if has_integrations then { integrations: this._integrations } else {}
    ),

    etc_config:: if has_prometheus_config then this.config {
      // Hide loki and integrations from our extra configs, we just want the
      // scrape configs that wouldn't work for the DaemonSet.
      prometheus+: {
        configs: std.map(function(cfg) cfg { host_filter: false }, etc_instances),
      },
      loki:: {},
      tempo:: {},
      integrations:: {},
    },

    agent:
      agent.newAgent(name, namespace, self._images.agent, self.config, use_daemonset=true) +
      agent.withConfigHash(self._config_hash) + {
        // Required for the scraping service; get the node name and store it in
        // $HOSTNAME so host_filtering works.
        container+:: container.withEnvMixin([
          k.core.v1.envVar.fromFieldPath('HOSTNAME', 'spec.nodeName'),
        ]),

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
      } + (
        if has_loki_config then $.lokiPermissionsMixin else {}
      ) + (
        if has_integrations && std.objectHas(this._integrations, 'node_exporter') then $.integrationsMixin else {}
      ),

    agent_etc: if std.length(etc_instances) > 0 then
      agent.newAgent(name + '-etc', namespace, self._images.agent, self.etc_config, use_daemonset=false) +
      agent.withConfigHash(self._config_hash),
  },

  // withImages sets the images used for launching the Agent.
  // Keys supported: agent, agentctl
  withImages(images):: { _images+: images },

  // Includes or excludes the config hash annotation.
  withConfigHash(include=true):: { _config_hash:: include },

  // withPortsMixin adds extra ports to expose.
  withPortsMixin(ports=[]):: {
    agent+: {
      container+:: container.withPortsMixin(ports),
    },
  },
}
