local agent = import 'grafana-agent/grafana-agent.libsonnet';

agent {
  _config+:: {
    namespace: 'default',
    agent_remote_write: [{
      url: '${REMOTE_WRITE_URL}',
      basic_auth: {
        username: '${REMOTE_WRITE_USERNAME}',
        password: '${REMOTE_WRITE_PASSWORD}',
      },
    }],
  },
}
