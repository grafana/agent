local agent = import 'grafana-agent/v1/main.libsonnet';

{
  agent:
    agent.newDeployment('grafana-agent', '${NAMESPACE}') +
    agent.withConfigHash(false) +
    agent.withImages({
      agent: (import 'version.libsonnet'),
    }) + {
      agent+: {
        // The listen port from the cloud config is set to 12345.
        listen_port:: 12345,

        // The bare deployment doesn't provide a ConfigMap by default, so
        // remove it from here.
        config_map:: {},
      },
    },
}
