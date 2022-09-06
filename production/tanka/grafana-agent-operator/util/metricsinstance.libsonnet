local gen = import 'agent-operator-gen/main.libsonnet';
local mi = gen.monitoring.v1alpha1.metricsInstance;
local rw = mi.spec.remoteWrite;

{
    withRemoteWrite(secretName, metricsUrl)::
        mi.spec.withRemoteWrite(
            rw.withUrl(metricsUrl) +
            rw.basicAuth.username.withKey('username') +
            rw.basicAuth.username.withName(secretName) +
            rw.basicAuth.password.withKey('password') +
            rw.basicAuth.password.withName(secretName)
        ),

    withNilServiceMonitorNamespace():: {
        spec+: {
            serviceMonitorNamespaceSelector: {}
        }
    }
}
