local agent = import './internal/agent.libsonnet';
local utils = import './internal/utils.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;
local configMap = k.core.v1.configMap;
local service = k.core.v1.service;

// Merge all of our libraries to create the final exposed library.
(import './lib/deployment.libsonnet') +
(import './lib/integrations.libsonnet') +
(import './lib/metrics.libsonnet') +
(import './lib/scraping_service.libsonnet') +
(import './lib/logs.libsonnet') +
(import './lib/traces.libsonnet') +
{
  _images:: {
    agent: 'grafana/agent:v0.34.0-rc.2',
    agentctl: 'grafana/agentctl:v0.34.0-rc.2',
  },

  // new creates a new DaemonSet deployment of the grafana-agent. By default,
  // the deployment will do no collection. You must merge the result of this
  // function with one or more of the following:
  //
  // - withMetricsConfig, withMetricsInstances (and optionally withRemoteWrite)
  // - withLogsConfig
  //
  // When using withMetricsInstances, a [name]-etc deployment
  // with one replica will be created alongside the DaemonSet. This deployment
  // is responsible for handling scrape configs that will not work on the host
  // machine.
  //
  // For example, if a scrape_config scrapes the Kubernetes API, that must be
  // handled by the [name]-etc deployment as the Kubernetes API does not run
  // on any node in the cluster.
  //
  // scrapeInstanceKubernetes provides the default
  // MetricsInstanceConfig Grafana Labs uses in production.
  new(name='grafana-agent', namespace='default'):: {
    local this = self,

    _mode:: 'daemonset',
    _images:: $._images,
    _config_hash:: true,

    local has_logs_config = std.objectHasAll(self, '_logs_config'),
    local has_trace_config = std.objectHasAll(self, '_trace_config'),
    local has_metrics_config = std.objectHasAll(self, '_metrics_config'),
    local has_metrics_instances = std.objectHasAll(self, '_metrics_instances'),
    local has_integrations = std.objectHasAll(self, '_integrations'),
    local has_sampling_strategies = std.objectHasAll(self, '_traces_sampling_strategies'),

    local metrics_instances =
      if has_metrics_instances then this._metrics_instances else [],
    local host_filter_instances = utils.transformInstances(metrics_instances, true),
    local etc_instances = utils.transformInstances(metrics_instances, false),

    config:: {
      server: {
        log_level: 'info',
      },
    } + (
      if has_metrics_config
      then { metrics: this._metrics_config { configs: host_filter_instances } }
      else {}
    ) + (
      if has_logs_config then {
        logs: {
          positions_directory: '/tmp/positions',
          configs: [this._logs_config {
            name: 'default',
          }],
        },
      } else {}
    ) + (
      if has_trace_config then {
        traces: {
          configs: [this._trace_config {
            name: 'default',
          }],
        },
      }
      else {}
    ) + (
      if has_integrations then { integrations: this._integrations } else {}
    ),

    etc_config:: if has_metrics_config then this.config {
      // Hide logs and integrations from our extra configs, we just want the
      // scrape configs that wouldn't work for the DaemonSet.
      metrics+: {
        configs: std.map(function(cfg) cfg { host_filter: false }, etc_instances),
      },
      logs:: {},
      traces:: {},
      integrations:: {},
    },

    agent:
      agent.newAgent(name, namespace, self._images.agent, self.config, use_daemonset=true) +
      agent.withConfigHash(self._config_hash) + {
        // If sampling strategies were defined, we need to mount them as a JSON
        // file.
        config_map+:
          if has_sampling_strategies
          then configMap.withDataMixin({
            'strategies.json': std.toString(this._traces_sampling_strategies),
          })
          else {},

        // If we're deploying for tracing, applications will want to write to
        // a service for load balancing span delivery.
        service:
          if has_trace_config
          then k.util.serviceFor(self.agent) + service.mixin.metadata.withNamespace(namespace)
          else {},
      } + (
        if has_logs_config then $.logsPermissionsMixin else {}
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
