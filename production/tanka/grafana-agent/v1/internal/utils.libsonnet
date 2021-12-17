{
  // Returns true if the scrape_config only contains a service_discovery for
  // Kubernetes (via kubernetes_sd_configs) that has role: pod
  isOnlyK8sPodDiscovery(scrape_config)::
    // Get all the *_sd_configs and filter that down to the sd_configs that aren't
    // kubernetes_sd_configs. It should be 0.
    std.length(std.filter(
      function(key) key != 'kubernetes_sd_configs',
      std.filter(
        function(key) std.endsWith(key, '_sd_configs'),
        std.objectFields(scrape_config),
      ),
    )) == 0 &&
    // Make sure there are 0 kubernetes_sd_configs whose role is not pod
    std.length(std.filter(
      function(kube_sd_config) kube_sd_config.role != 'pod',
      std.flatMap(
        function(key) scrape_config[key],
        std.filter(
          function(key) key == 'kubernetes_sd_configs',
          std.objectFields(scrape_config)
        )
      )
    )) == 0,

  // host_filter_compatible instances are ones that:
  // - only use kubernetes_sd_configs
  // - only use kubernetes_sd_configs with role = 'pod'
  transformInstances(instances=[], host_filter_compatible=true)::
    std.map(function(instance) instance {
      scrape_configs: std.filter(
        function(cfg) $.isOnlyK8sPodDiscovery(cfg) == host_filter_compatible,
        super.scrape_configs,
      ),
    }, instances),
}
