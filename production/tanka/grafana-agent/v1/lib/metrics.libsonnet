local scrape_k8s = import '../internal/kubernetes_instance.libsonnet';

{
  // defaultMetricsConfig holds the default Metrics Config with all
  // options that the Agent supports. It is better to use this object as a
  // reference rather than extending it; since all fields are defined here, if
  // the Agent changes a default value in the future, the default change will
  // be overridden by the values here.
  //
  // Required fields will be marked as REQUIRED.
  defaultMetricsConfig:: {
    // Settings that apply to all launched Metrics instances by default.
    // These settings may be overridden on a per-instance basis.
    global: {
      // How frequently to scrape for metrics.
      scrape_interval: '1m',

      // How long to wait before timing out from scraping a target.
      scrape_timeout: '10s',

      // Extra labels to apply to all scraped targets.
      external_labels: {
        /* foo: 'bar', */
      },
    },

    // Where to store the WAL for metrics before they are sent to remote_write.
    // REQUIRED. The value here is preconfigured to work with the Tanka configs.
    wal_directory: '/var/lib/agent/data',

    // If an instance crashes abnormally, wait this long before restarting it.
    // 0s disables the backoff period and restarts the instance immediately.
    instance_restart_backoff: '5s',

    // How to spawn instances based on compatible fields. Supported values:
    // "shared" (default), "distinct".
    instance_mode: 'shared',
  },

  // withMetricsConfig controls the Metrics engine settings for the Agent.
  // defaultMetricsConfig explicitly defines all supported values that can be
  // provided within config.
  withMetricsConfig(config):: { _metrics_config:: config },

  // withMetricsInstances controls the Metrics instances the Agent will
  // launch. Instances may be a single object or an array of objects. Each
  // object must have a name key that is unique to that object.
  //
  // scrapeInstanceKubernetes defines an example set of instances and the
  // ones Grafana Labs uses in production. It does not demonstrate all available
  // values for scrape configs and remote_write. For detailed information on
  // instance config settings, consult the Agent documentation:
  //
  // https://github.com/grafana/agent/blob/main/docs/configuration-reference.md#metrics_instance_config
  //
  // host_filter does not need to be applied here; the library will apply it
  // automatically based on how the Agent is being deployed.
  //
  // remote_write rules may be defined in the instance object. Optionally,
  // remove_write rules may be applied to every instance object by using the
  // withRemoteWrite function.
  withMetricsInstances(instances):: {
    assert std.objectHasAll(self, '_mode') : |||
      withMetricsInstances must be merged with the result of calling new,
      newDeployment, or newScrapingService.
    |||,

    local list = if std.isArray(instances) then instances else [instances],

    // If the library was invoked in daemonset mode, we want to use
    // host_filtering mode so each Agent only scrapes stuff from its local
    // machine.
    local host_filter = super._mode == 'daemonset',

    // Apply host_filtering over our list of instances.
    _metrics_instances:: std.map(function(inst) inst {
      host_filter: host_filter,

      // Make sure remote_write is an empty array if it doesn't exist.
      remote_write:
        if !std.objectHas(inst, 'remote_write') || !std.isArray(inst.remote_write)
        then []
        else inst.remote_write,
    }, list),
  },

  // withRemoteWrite overwrites all the remote_write configs provided in
  // withMetricsInstances with the specified remote_writes. This is
  // useful when there are multiple instances and you just want everything
  // to remote_write to the same place.
  //
  // Refer to the remote_write specification for all available fields:
  //   https://github.com/grafana/agent/blob/main/docs/configuration-reference.md#remote_write
  withRemoteWrite(remote_writes):: {
    assert std.objectHasAll(self, '_mode') : |||
      withMetricsInstances must be merged with the result of calling new,
      newDeployment, or newScrapingService.
    |||,

    local list = if std.isArray(remote_writes) then remote_writes else [remote_writes],
    _metrics_config+:: { global+: { remote_write: list } },
  },

  // scrapeInstanceKubernetes defines an instance config Grafana Labs uses to
  // scrape Kubernetes metrics.
  //
  // Pods will be scraped if:
  //
  // 1. They have a port ending in -metrics
  // 2. They do not have a prometheus.io/scrape=false annotation
  // 3. They have a name label
  scrapeInstanceKubernetes: scrape_k8s.newKubernetesScrapeInstance(
    config=scrape_k8s.kubernetesScrapeInstanceConfig,
    namespace='default',
  ),
}
