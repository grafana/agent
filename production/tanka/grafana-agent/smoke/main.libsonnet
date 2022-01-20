local smoke = import './internal/smoke.libsonnet';

{
    _images:: {
        agentsmoke: 'us.gcr.io/kubernetes-dev/grafana/agent-smoke:latest',
    },

    new(name='grafana-agent-smoke', namespace='grafana-agent-smoke', mutationFrequency='5m', chaosFrequency='30m', image=self._images.agentsmoke):: {
        smoke:
            smoke.newSmoke(name, namespace, mutationFrequency, chaosFrequency, image)
    },

    monitoring: (import './prometheus_monitoring.libsonnet'),
}
