{
  local d = (import 'doc-util/main.libsonnet'),
  '#':: d.pkg(name='v1alpha1', url='', help=''),
  grafanaAgent: (import 'grafanaAgent.libsonnet'),
  prometheusInstance: (import 'prometheusInstance.libsonnet'),
}
