{
  local version = std.extVar('BUILD_IMAGE_VERSION'),

  //linux: 'grafana/agent-build-image:%s' % version,
  windows: 'grafana/agent-build-image:%s-windows' % version,

  linux: 'rfratto/agent-build-image:0.21.0-go-patch-5',
}
