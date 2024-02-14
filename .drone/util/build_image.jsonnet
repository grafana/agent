{
  local version = std.extVar('BUILD_IMAGE_VERSION'),

  linux: 'grafana/agent-build-image:%s' % version,
  windows: 'grafana/agent-build-image:%s-windows' % version,
  boringcrypto: 'grafana/agent-build-image:%s-boringcrypto' % version,
}
