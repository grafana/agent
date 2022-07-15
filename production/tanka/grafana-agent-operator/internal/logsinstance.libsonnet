function(name='grafana-agent-logs', namespace='', external_labels={}) {    
    local this = self,
    local gen = import 'agent-operator-gen/main.libsonnet',
    local k = import 'ksonnet-util/kausal.libsonnet',
    
    local secret = k.core.v1.secret,
    local li = gen.monitoring.v1alpha1.logsInstance,
    local ga = gen.monitoring.v1alpha1.grafanaAgent,
    local clients = li.spec.clients,

    _config+:: {
        def_li_labels: {agent: "grafana-agent-logs"},
        logs_rw_url: 'YOUR_HL_URL',
        logs_rw_user: 'YOUR_HL_USER',
        logs_rw_pass: 'YOUR_HL_PASS',
        logs_secret_name: 'primary-credentials-logs'
    },

    // todo: can make this cleaner?
    logs_client::
        clients.withUrl(this._config.logs_rw_url) +
        clients.basicAuth.username.withKey('username') +
        clients.basicAuth.username.withName(this._config.logs_secret_name) +
        clients.basicAuth.password.withKey('password') +
        clients.basicAuth.password.withName(this._config.logs_secret_name) +
        clients.withExternalLabels(external_labels),

    // todo: kill nil data field
    logs_secret: secret.new(this._config.logs_secret_name, {}) +
    secret.withStringData({
        username: this._config.logs_rw_user,
        password: this._config.logs_rw_pass,
    }) +
    secret.mixin.metadata.withNamespace(namespace),

    logs_instance: li.new(name) +
        li.metadata.withNamespace(namespace) +
        li.metadata.withLabels(this._config.def_li_labels) +
        li.spec.withClients(this.logs_client),

    ga_resource+: ga.spec.logs.instanceSelector.withMatchLabels(this._config.def_li_labels)

}
