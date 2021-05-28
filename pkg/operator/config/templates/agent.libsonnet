// agent.libsonnet is the entrypoint for rendering a Grafana Agent config file
// based on the Operator custom resources.
//
// When writing an object, any field will null will be removed from the final
// YAML. This is useful as we don't want to always translate unfilled values
// from the custom resources to a field in the YAML.
//
// A series of helper methods to convert default values into null (so they can
// be trimmed) are in ./ext/optionals.libsonnet.
//
// When writing a new function, please document the expected types of the
// arguments.

local marshal = import './ext/marshal.libsonnet';
local optionals = import './ext/optionals.libsonnet';

local new_external_labels = import './component/external_labels.libsonnet';
local new_remote_write = import './component/remote_write.libsonnet';
local new_prometheus_instance = import './prometheus.libsonnet';

local calculateShards(requested) =
  if requested == null then 1
  else if requested > 1 then requested
  else 1;

// @param {config.Deployment} ctx
function(ctx) marshal.YAML(optionals.trim({
  local spec = ctx.Agent.Spec,
  local prometheus = spec.Prometheus,
  local namespace = ctx.Agent.ObjectMeta.Namespace,

  server: {
    http_listen_port: 8080,
    log_level: optionals.string(spec.LogLevel),
    log_format: optionals.string(spec.LogFormat),
  },

  prometheus: {
    wal_directory: '/var/lib/grafana-agent/data',
    global: {
      external_labels: optionals.object(new_external_labels(ctx)),
      scrape_interval: optionals.string(prometheus.ScrapeInterval),
      scrape_timeout: optionals.string(prometheus.ScrapeTimeout),
      remote_write: optionals.array(std.map(
        function(rw) new_remote_write(ctx.Agent.ObjectMeta.Namespace, rw),
        prometheus.RemoteWrite,
      )),
    },
    configs: optionals.array(std.map(
      function(inst) new_prometheus_instance(
        agentNamespace=ctx.Agent.ObjectMeta.Namespace,
        instance=inst,
        apiServer=spec.APIServerConfig,
        overrideHonorLabels=prometheus.OverrideHonorLabels,
        overrideHonorTimestamps=prometheus.OverrideHonorTimestamps,
        ignoreNamespaceSelectors=prometheus.IgnoreNamespaceSelectors,
        enforcedNamespaceLabel=prometheus.EnforcedNamepsaceLabel,
        enforcedSampleLimit=prometheus.EnforcedSampleLimit,
        enforcedTargetLimit=prometheus.EnforcedTargetLimit,
        shards=calculateShards(prometheus.Shards),
      ),
      ctx.Prometheis,
    )),
  },
}))
