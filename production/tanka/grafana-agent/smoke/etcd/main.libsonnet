local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;
local containerPort = k.core.v1.containerPort;
local deployment = k.apps.v1.deployment;
local service = k.core.v1.service;

{
  new(namespace=''):: {
    container::
      container.new('etcd', 'gcr.io/etcd-development/etcd:v3.4.7') +
      container.withPorts([
        containerPort.newNamed(name='etcd', containerPort=2379),
      ]) +
      container.withArgsMixin([
        '/usr/local/bin/etcd',
        '--listen-client-urls=http://0.0.0.0:2379',
        '--advertise-client-urls=http://0.0.0.0:2379',
        '--log-level=info',
      ]),

    deployment:
      deployment.new('etcd', 1, [self.container]) +
      deployment.mixin.metadata.withNamespace(namespace),

    service:
      k.util.serviceFor(self.deployment) +
      service.mixin.metadata.withNamespace(namespace),
  },
}
