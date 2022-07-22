local gen = import 'agent-operator-gen/main.libsonnet';
local li = gen.monitoring.v1alpha1.logsInstance;
local clients = li.spec.clients;

{
    withLogsClient(secretName, logsUrl, externalLabels={})::
        li.spec.withClients(
            clients.withUrl(logsUrl) +
            clients.basicAuth.username.withKey('username') +
            clients.basicAuth.username.withName(secretName) +
            clients.basicAuth.password.withKey('password') +
            clients.basicAuth.password.withName(secretName) + 
            if externalLabels != {} then clients.withExternalLabels(externalLabels) else {}
        ),

    withNilPodLogsNamespace():: {
        spec+: {
            podLogsNamespaceSelector: {}
        }
    },
}
