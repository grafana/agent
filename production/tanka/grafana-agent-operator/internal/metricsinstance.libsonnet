function(name='grafana-agent-metrics', namespace='', config={}) {    
    local this = self,
    local gen = import 'agent-operator-gen/main.libsonnet',
    local k = import 'ksonnet-util/kausal.libsonnet',
    
    local secret = k.core.v1.secret,

    local mi = gen.monitoring.v1alpha1.metricsInstance,
    local ga = gen.monitoring.v1alpha1.grafanaAgent,
    local rw = mi.spec.remoteWrite,

    _config:: {
        mi_labels: {agent: "grafana-agent-metrics"},
        metrics_rw_url: 'YOUR_HM_URL',
        metrics_rw_user: 'YOUR_HM_USER',
        metrics_rw_pass: 'YOUR_HM_PASS',
        metrics_secret_name: 'primary-credentials-metrics',
        external_labels: {cluster: 'cloud'},
    } + config,

    local metricsRemoteWrite() =
        rw.withUrl(this._config.metrics_rw_url) +
        rw.basicAuth.username.withKey('username') +
        rw.basicAuth.username.withName(this._config.metrics_secret_name) +
        rw.basicAuth.password.withKey('password') +
        rw.basicAuth.password.withName(this._config.metrics_secret_name),

    // todo(hjet): kill nil data field
    secret: secret.new(this._config.metrics_secret_name, {}) +
    secret.withStringData({
        username: this._config.metrics_rw_user,
        password: this._config.metrics_rw_pass,
    }) +
    secret.mixin.metadata.withNamespace(namespace),

    resource: mi.new(name) +
        mi.metadata.withNamespace(namespace) +
        mi.metadata.withLabels(this._config.mi_labels) +
        mi.spec.withRemoteWrite(metricsRemoteWrite()),
}
