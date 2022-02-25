local agent = import 'grafana-agent/v1/main.libsonnet';

{
  agent:
    agent.new('grafana-agent-logs', '${NAMESPACE}') +
    agent.withConfigHash(false) +
    agent.withImages({
      agent: (import 'version.libsonnet'),
    }) + agent.withLogsConfig(agent.scrapeKubernetesLogs) + {
      agent+: {
        // Remove ConfigMap
        config_map:: {},
      },
    },
}
