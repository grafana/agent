local gen = import 'agent-operator-gen/main.libsonnet';
local li = gen.monitoring.v1alpha1.logsInstance;
local k = import 'ksonnet-util/kausal.libsonnet';
local secret = k.core.v1.secret;
local clients = li.spec.clients;

{
    new(name='grafana-agent-logs', namespace='', logsInstanceLabels):: {    
        resource: li.new(name) +
            li.metadata.withNamespace(namespace) +
            li.metadata.withLabels(logsInstanceLabels)
    },

    withLogsClient(namespace, secretName, logsUrl, logsUser, logsKey, externalLabels={}):: {
        secret: 
            secret.new(secretName, {}) +
            secret.withStringData({
                username: logsUser,
                password: logsKey,
            }) + secret.mixin.metadata.withNamespace(namespace),

        resource+:
            li.spec.withClients(
                clients.withUrl(logsUrl) +
                clients.basicAuth.username.withKey('username') +
                clients.basicAuth.username.withName(secretName) +
                clients.basicAuth.password.withKey('password') +
                clients.basicAuth.password.withName(secretName) + 
                if externalLabels != {} then clients.withExternalLabels(externalLabels) else {}
            ),
    },

    withPodLogs(podLogsLabels):: {
        resource+:
            li.spec.podLogsSelector.withMatchLabels(podLogsLabels)
    },

    withNilPodLogsNamespace():: {
        resource+: {
            spec+: {
                podLogsNamespaceSelector: {}
        }
    },
  },
}
