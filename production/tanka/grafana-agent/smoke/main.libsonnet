local smoke = import './internal/smoke.libsonnet';

{
    _images:: {
        agentsmoke: 'us.gcr.io/kubernetes-dev/grafana/agent-smoke:latest',
    },

    new(name='grafana-agent-smoke', namespace='grafana-agent-smoke', image=self._images.agentsmoke):: {
        local this = self,

        smoke:
            smoke.newSmoke(name, namespace, image)
    }

}