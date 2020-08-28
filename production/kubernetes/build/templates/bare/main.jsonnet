local agent = import 'grafana-agent/grafana-agent.libsonnet';

agent {
  _images+:: {
    agent: (import 'version.libsonnet'),
  },

  _config+:: {
    namespace: 'default',

    // Since the config map isn't managed by Tanka, we don't want to
    // add the configmap's hash as an annotation for the Kubernetes
    // YAML manifest.
    agent_config_hash_annotation: false,
  },

  // We're describing a deployment where the ConfigMap isn't available, 
  // so remove the various config maps and components that aren't relevant.
  agent_config_map:: {},
  agent_deployment_config_map:: {},
  agent_deployment:: {},
}
