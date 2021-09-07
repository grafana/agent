local k8s_v2 = import '../../v2/internal/helpers/k8s.libsonnet';

{
  newKubernetesLogsCollector():: {
    scrape_configs: k8s_v2.logs(),
  },
}
