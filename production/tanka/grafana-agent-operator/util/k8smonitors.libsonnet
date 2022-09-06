local prom_gen = import 'prom-operator-gen/main.libsonnet';
local sm = prom_gen.monitoring.v1.serviceMonitor;
local e = sm.spec.endpoints;
local mr = e.metricRelabelings;
local r = e.relabelings;

{
    local metricArrayToString(arr) = std.join("|", arr),

    local withJobReplace(job_label) =
        r.withAction('replace') +
        r.withTargetLabel('job') +
        r.withReplacement(job_label),
        
    local withAllowList(metrics) =
        mr.withAction('keep') +
        mr.withSourceLabels(['__name__']) +
        mr.withRegex(metricArrayToString(metrics)),

    local withMetricsPath() =
        r.withSourceLabels(['__metrics_path__']) +
        r.withTargetLabel('metrics_path'),
        
    local withDefaultEndpoint(jobLabel, port, allowlist, allowlistMetrics, path) =
        e.withHonorLabels(true) +
        e.withInterval('60s') +
        (if allowlist then e.withMetricRelabelings(withAllowList(allowlistMetrics)) else {}) +
        e.withPort(port) +
        e.withPath(path),
    
    
    newKubernetesMonitor(name, namespace, monitorLabels, targetNamespace, targetLabels, jobLabel, metricsPath, allowlist=false, allowlistMetrics=[])::
        sm.new(name) +
        sm.metadata.withNamespace(namespace) +
        sm.metadata.withLabels(monitorLabels) +
        sm.spec.namespaceSelector.withMatchNames(targetNamespace) +
        sm.spec.selector.withMatchLabels(targetLabels) +
        sm.spec.withEndpoints([
            withDefaultEndpoint(jobLabel, 'https-metrics', allowlist, allowlistMetrics, metricsPath) +
            e.withBearerTokenFile('/var/run/secrets/kubernetes.io/serviceaccount/token') +
            e.tlsConfig.withInsecureSkipVerify(true) +
            e.withRelabelings([withMetricsPath(), withJobReplace(jobLabel)]) +
            e.withScheme('https')
        ]),
    
    newServiceMonitor(name, namespace, monitorLabels, targetNamespace, targetLabels, jobLabel, metricsPath, allowlist=false, allowlistMetrics=[])::
        sm.new(name) +
        sm.metadata.withNamespace(namespace) +
        sm.metadata.withLabels(monitorLabels) +
        sm.spec.namespaceSelector.withMatchNames(targetNamespace) +
        sm.spec.selector.withMatchLabels(targetLabels) +
        sm.spec.withEndpoints([
            withDefaultEndpoint(jobLabel, 'http-metrics', allowlist, allowlistMetrics, metricsPath) +
            e.withRelabelings([withJobReplace(jobLabel)])
        ]),
}
