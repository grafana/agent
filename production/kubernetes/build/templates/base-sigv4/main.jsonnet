local agent = import 'grafana-agent/grafana-agent.libsonnet';

agent {
  _images+:: {
    agent: (import 'version.libsonnet'),
  },

  _config+:: {
    namespace: 'default',
    agent_remote_write: [{
      url: '${REMOTE_WRITE_URL}',
      sigv4: {
        enabled: true,
      },
    }],

    // Since the config map isn't managed by Tanka, we don't want to
    // add the configmap's hash as an annotation for the Kubernetes
    // YAML manifest.
    agent_config_hash_annotation: false,
  },
}
