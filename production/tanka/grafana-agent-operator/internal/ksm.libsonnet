function(name='kube-state-metrics', namespace='') {    
    local ksm = import 'kube-state-metrics/kube-state-metrics.libsonnet',

    kube_state_metrics:
        ksm {
            name:: name,
            namespace:: namespace,
            version:: '2.5.0',
            image:: 'registry.k8s.io/kube-state-metrics/kube-state-metrics:v' + self.version,
        }
}
