local gen = import 'agent-operator-gen/main.libsonnet';
local mi = gen.monitoring.v1alpha1.metricsInstance;
local rw = mi.spec.remoteWrite;
local k = import 'ksonnet-util/kausal.libsonnet';
local secret = k.core.v1.secret;

{   
    new(name='grafana-agent-metrics', namespace='', metricsInstanceLabels):: {
        resource:
            mi.new(name) +
            mi.metadata.withNamespace(namespace) +
            mi.metadata.withLabels(metricsInstanceLabels)
    },

    withRemoteWrite(namespace, secretName, metricsUrl, metricsUser, metricsKey):: {
        secret: 
            secret.new(secretName, {}) +
            secret.withStringData({
                username: metricsUser,
                password: metricsKey,
            }) + secret.mixin.metadata.withNamespace(namespace),

        resource+:
            mi.spec.withRemoteWrite(
                rw.withUrl(metricsUrl) +
                rw.basicAuth.username.withKey('username') +
                rw.basicAuth.username.withName(secretName) +
                rw.basicAuth.password.withKey('password') +
                rw.basicAuth.password.withName(secretName)
            ),
    },

    withServiceMonitor(serviceMonitorLabels):: {
        resource+:
            mi.spec.serviceMonitorSelector.withMatchLabels(serviceMonitorLabels)
    },

    withNilServiceMonitorNamespace():: {
        resource+: {
            spec+: {
                serviceMonitorNamespaceSelector: {}
            }
        }
    },
}
