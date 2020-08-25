local agent = import 'grafana-agent/v1/main.libsonnet';

{
  agent:
    agent.new() +
    agent.withImages({
      agent: (import 'version.libsonnet'),
      agentctl: 'grafana/agentctl:v0.5.0',
    }) +
    agent.withPrometheusConfig({
      wal_directory: '/var/lib/agent/data',
    }) +
    agent.withPrometheusInstances(agent.scrapeInstanceKubernetes) +
    agent.withRemoteWrite({
      basic_auth: {
        username: '${REMOTE_WRITE_USERNAME}',
        password: '${REMOTE_WRITE_PASSWORD}',
      },
    }) +
    // Since the config map isn't managed by Tanka, we don't want to
    // add the configmap's hash as an annotation for the Kubernetes
    // YAML manifest.
    agent.withConfigHash(false),
}
