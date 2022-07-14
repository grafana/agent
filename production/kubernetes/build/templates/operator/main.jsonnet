local k = import 'ksonnet-util/kausal.libsonnet';
local agent_operator = import 'grafana-agent-operator/main.libsonnet';

{
  agent_operator:
    agent_operator.new(name='grafana-agent-operator', namespace='default')
}
