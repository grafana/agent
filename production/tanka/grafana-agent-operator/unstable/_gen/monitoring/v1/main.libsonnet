{
  local d = (import 'doc-util/main.libsonnet'),
  '#':: d.pkg(name='v1', url='', help=''),
  podMonitor: (import 'podMonitor.libsonnet'),
  probe: (import 'probe.libsonnet'),
  serviceMonitor: (import 'serviceMonitor.libsonnet'),
}
