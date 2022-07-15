function(namespace='', allowlist=false, allowlistmetrics=[]) {    
    local this = self,
    local gen = import 'agent-operator-gen/main.libsonnet',
    local prom_gen = import 'prom-operator-gen/main.libsonnet',
    local k = import 'ksonnet-util/kausal.libsonnet',
    
    local mi = gen.monitoring.v1alpha1.metricsInstance,
    local sm = prom_gen.monitoring.v1.serviceMonitor,

    _config+:: {
        def_monitor_labels: {instance: "primary"},
        kubelet_svc_label: {'app.kubernetes.io/name': 'kubelet'}
    },

    local withNilServiceMonitorNamespaceSelector() = {
        spec+: {
            serviceMonitorNamespaceSelector: {}
        }
    },

    local withNilPodMonitorNamespaceSelector() = {
        spec+: {
            podMonitorNamespaceSelector: {}
        }
    },

    local withNilProbeNamespaceSelector() = {
        spec+: {
            probeNamespaceSelector: {}
        }
    },

    kubelet_monitor: sm.new('kubelet-monitor') +
        sm.metadata.withNamespace(namespace) +
        sm.metadata.withLabels(this._config.def_monitor_labels) +
        sm.spec.namespaceSelector.withAny(true) +
        sm.spec.selector.withMatchLabels(this._config.kubelet_svc_label),

    # TODO: cadvisor SM

    metrics_instance+: withNilServiceMonitorNamespaceSelector() +
        mi.spec.serviceMonitorSelector.withMatchLabels(this._config.def_monitor_labels) +
        withNilPodMonitorNamespaceSelector() +
        mi.spec.podMonitorSelector.withMatchLabels(this._config.def_monitor_labels) +
        withNilProbeNamespaceSelector() +
        mi.spec.probeSelector.withMatchLabels(this._config.def_monitor_labels),
}
