local agent = import 'grafana-agent/grafana-agent.libsonnet';

local k = import 'ksonnet-util/kausal.libsonnet';
local serviceAccount = k.core.v1.serviceAccount;

agent {
  _images+:: {
    agent: (import 'version.libsonnet'),
  },

  _config+:: {
    namespace: '${NAMESPACE}',
    agent_remote_write: [{
      url: '${REMOTE_WRITE_URL}',
      sigv4: {
        region: '${REGION}',
      },
    }],

    // Since the config map isn't managed by Tanka, we don't want to
    // add the configmap's hash as an annotation for the Kubernetes
    // YAML manifest.
    agent_config_hash_annotation: false,
  },

  agent_rbac+: {
    service_account+: serviceAccount.mixin.metadata.withAnnotationsMixin({
      'eks.amazonaws.com/role-arn': '${ROLE_ARN}',
    }),
  },
}
