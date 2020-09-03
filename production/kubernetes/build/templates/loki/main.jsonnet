local agent = import 'grafana-agent/v1/main.libsonnet';

{
  agent:
    agent.new('grafana-agent-logs', 'default') +
    agent.withConfigHash(false) +
    agent.withImages({
      agent: (import 'version.libsonnet'),
    }) +
    agent.withLokiConfig(agent.scrapeKubernetesLogs) +
    agent.withLokiClients(
      agent.newLokiClient({
        scheme: 'https',
        hostname: '${LOKI_HOSTNAME}',
        username: '${LOKI_USERNAME}',
        password: '${LOKI_PASSWORD}',
      })
    ),
}
