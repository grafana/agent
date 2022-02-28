local agent = import 'grafana-agent/v2/main.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

{
  agent:
    agent.new(name='grafana-agent-logs', namespace='${NAMESPACE}') + 
    agent.withDaemonSetController() + 
    agent.withConfigHash(false) +
    // add dummy config or else will fail
    agent.withAgentConfig({
      server: { log_level: 'error' },
    }) + 
    agent.withLogVolumeMounts(config={}) +
    agent.withLogPermissions(config={}) +
    // hack to disable configmap
    { configMap:: super.configMap }
}
