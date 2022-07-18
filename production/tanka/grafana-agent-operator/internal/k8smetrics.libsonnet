function(namespace='', allowlist=false, allowlistmetrics=[], config) {    
    local this = self,
    local gen = import 'agent-operator-gen/main.libsonnet',
    local prom_gen = import 'prom-operator-gen/main.libsonnet',
    local k = import 'ksonnet-util/kausal.libsonnet',
    
    local mi = gen.monitoring.v1alpha1.metricsInstance,
    local sm = prom_gen.monitoring.v1.serviceMonitor,
    local e = sm.spec.endpoints,
    local mr = e.metricRelabelings,
    local r = e.relabelings,
    
    _config:: {
        monitor_labels: {agent: 'grafana-agent-metrics'},
        kubelet_svc_label: {'app.kubernetes.io/name': 'kubelet'},
        ksm_svc_label: {'app.kubernetes.io/name': 'kube-state-metrics'}
    } + config,

    local metricArrayToString(arr) = std.join("|", arr),

    local withJobReplace(job_label) =
        mr.withAction('replace') +
        mr.withTargetLabel('job') +
        mr.withReplacement(job_label),
    
    local withAllowList(metrics) =
        mr.withAction('keep') +
        mr.withSourceLabels(['__name__']) +
        mr.withRegex(metricArrayToString(metrics)),

    local withMetricsPath() =
        r.withSourceLabels(['__metrics_path__']) +
        r.withTargetLabel('metrics_path'),
    
    local withDefaultEndpoint(job_label, port, path='/metrics') =
        e.withHonorLabels(true) +
        e.withInterval('60s') +
        e.withMetricRelabelings(if allowlist then [withJobReplace(job_label), withAllowList(allowlistmetrics)] else [withJobReplace(job_label)]) +
        e.withPort(port) +
        e.withPath(path),
    
    kubelet_monitor: sm.new('kubelet-monitor') +
        sm.metadata.withNamespace(namespace) +
        sm.metadata.withLabels(this._config.monitor_labels) +
        sm.spec.namespaceSelector.withAny(true) +
        sm.spec.selector.withMatchLabels(this._config.kubelet_svc_label) +
        sm.spec.withEndpoints([
            withDefaultEndpoint('integrations/kubernetes/kubelet', 'https-metrics') +
            e.withBearerTokenFile('/var/run/secrets/kubernetes.io/serviceaccount/token') +
            e.tlsConfig.withInsecureSkipVerify(true) +
            e.withRelabelings([withMetricsPath()]) +
            e.withScheme('https')
        ]),

    cadvisor_monitor: sm.new('cadvisor-monitor') +
        sm.metadata.withNamespace(namespace) +
        sm.metadata.withLabels(this._config.monitor_labels) +
        sm.spec.namespaceSelector.withAny(true) +
        sm.spec.selector.withMatchLabels(this._config.kubelet_svc_label) +
        sm.spec.withEndpoints([
            withDefaultEndpoint('integrations/kubernetes/cadvisor', 'https-metrics', '/metrics/cadvisor') +
            e.withBearerTokenFile('/var/run/secrets/kubernetes.io/serviceaccount/token') +
            e.tlsConfig.withInsecureSkipVerify(true) +
            e.withRelabelings([withMetricsPath()]) +
            e.withScheme('https')
        ]),

    ksm_monitor: sm.new('ksm-monitor') +
        sm.metadata.withLabels(this._config.monitor_labels) +
        sm.spec.namespaceSelector.withAny(true) +
        sm.spec.selector.withMatchLabels(this._config.ksm_svc_label) +
        sm.spec.withEndpoints([
            withDefaultEndpoint('integrations/kubernetes/kube-state-metrics', 'http-metrics')
        ]),
}
