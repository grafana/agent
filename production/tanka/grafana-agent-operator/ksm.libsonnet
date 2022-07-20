local ksm = import 'kube-state-metrics/kube-state-metrics.libsonnet';

{
    new(name, namespace):: {    
        kube_state_metrics:
            ksm {
                name:: name,
                namespace:: namespace,
                version:: '2.5.0',
                image:: 'registry.k8s.io/kube-state-metrics/kube-state-metrics:v' + self.version,
            }
    }
}
