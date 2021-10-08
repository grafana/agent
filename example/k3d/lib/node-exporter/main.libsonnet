local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local daemonSet = k.apps.v1.daemonSet;
local service = k.core.v1.service;

{
  new(namespace=''):: {
    container::
      container.new('node-exporter', 'quay.io/prometheus/node-exporter:v1.1.2') +
      container.withPorts([
        containerPort.newNamed(name='http-metrics', containerPort=9100),
      ]) +
      container.withArgsMixin([
        '--path.rootfs=/host/root',
        '--path.procfs=/host/proc',
        '--path.sysfs=/host/sys',
        '--collector.netdev.device-exclude=^veth.+$',
      ]) +
      container.mixin.securityContext.withPrivileged(true) +
      container.mixin.securityContext.withRunAsUser(0),

    daemonSet:
      daemonSet.new('node-exporter', [self.container]) +
      daemonSet.mixin.metadata.withNamespace(namespace) +
      daemonSet.mixin.spec.template.metadata.withAnnotationsMixin({ 'prometheus.io.scrape': 'false' }) +
      daemonSet.mixin.spec.template.spec.withHostPid(true) +
      daemonSet.mixin.spec.template.spec.withHostNetwork(true) +
      k.util.hostVolumeMount('proc', '/proc', '/host/proc') +
      k.util.hostVolumeMount('sys', '/sys', '/host/sys') +
      k.util.hostVolumeMount('root', '/', '/host/root') +
      k.util.podPriority('critical'),
  },
}
