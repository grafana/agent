function(name='grafana-agent-metrics', namespace='', external_labels={}) {    
    local this = self,
    local gen = import 'agent-operator-gen/main.libsonnet',
    local k = import 'ksonnet-util/kausal.libsonnet',
    
    local secret = k.core.v1.secret,
    local mi = gen.monitoring.v1alpha1.metricsInstance,
    local ga = gen.monitoring.v1alpha1.grafanaAgent,
    local remote_write = mi.spec.remoteWrite,

    _config+:: {
        def_mi_labels: {agent: "grafana-agent-metrics"},
        metrics_rw_url: 'YOUR_HM_URL',
        metrics_rw_user: 'YOUR_HM_USER',
        metrics_rw_pass: 'YOUR_HM_PASS',
        metrics_secret_name: 'primary-credentials-metrics'
    },

    // todo: can make this cleaner?
    metrics_rw::
        remote_write.withUrl(this._config.metrics_rw_url) +
        remote_write.basicAuth.username.withKey('username') +
        remote_write.basicAuth.username.withName(this._config.metrics_secret_name) +
        remote_write.basicAuth.password.withKey('password') +
        remote_write.basicAuth.password.withName(this._config.metrics_secret_name),

    // todo: kill nil data field
    metrics_secret: secret.new(this._config.metrics_secret_name, {}) +
    secret.withStringData({
        username: this._config.metrics_rw_user,
        password: this._config.metrics_rw_pass,
    }) +
    secret.mixin.metadata.withNamespace(namespace),

    metrics_instance: mi.new(name) +
        mi.metadata.withNamespace(namespace) +
        mi.metadata.withLabels(this._config.def_mi_labels) +
        mi.spec.withRemoteWrite(this.metrics_rw),

    // link back to parent to pick this MI up
    ga_resource+: ga.spec.metrics.withExternalLabels(external_labels) + ga.spec.metrics.instanceSelector.withMatchLabels(this._config.def_mi_labels)
}
