local agent = import 'grafana-agent/v1/main.libsonnet';

{
  local deployment =
    agent.new() +
    agent.withImages({
      agent: (import 'version.libsonnet'),
      agentctl: 'grafana/agentctl:v0.5.0',
    }) +
    // Since the config map isn't managed by Tanka, we don't want to
    // add the configmap's hash as an annotation for the Kubernetes
    // YAML manifest.
    agent.withConfigHash(false),

  // The bare deployment doesn't have the config map, so we'll 
  // hack into the internal state and remove it here.
  agent: deployment {
    agent+: {
      config_map:: {},
    },
  },
}
