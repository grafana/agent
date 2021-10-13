local k8s_v2 = import '../../v2/internal/helpers/k8s.libsonnet';

{
  kubernetesScrapeInstanceConfig:: {
    scrape_api_server_endpoints: false,
    insecure_skip_verify: false,

    cluster_dns_tld: 'local',
    cluster_dns_suffix: 'cluster.' + self.cluster_dns_tld,
    kubernetes_api_server_address: 'kubernetes.default.svc.%(cluster_dns_suffix)s:443' % self,
  },

  newKubernetesScrapeInstance(config, namespace='default'):: {
    local _config = $.kubernetesScrapeInstanceConfig + config,

    name: 'kubernetes',
    scrape_configs: k8s_v2.metrics({
      scrape_api_server_endpoints: _config.scrape_api_server_endpoints,
      insecure_skip_verify: _config.insecure_skip_verify,
      cluster_dns_tld: _config.cluster_dns_tld,
      cluster_dns_suffix: _config.cluster_dns_suffix,
      kubernetes_api_server_address: _config.kubernetes_api_server_address,
      ksm_namespace: namespace,
      node_exporter_namespace: namespace,
    }),
  },
}
