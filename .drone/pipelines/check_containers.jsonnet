local build_image = import '../util/build_image.jsonnet';
local pipelines = import '../util/pipelines.jsonnet';

local linux_containers = [
  { name: 'grafana/agent', make: 'make agent-image', path: 'cmd/grafana-agent/Dockerfile' },
  { name: 'grafana/agentctl', make: 'make agentctl-image', path: 'cmd/grafana-agentctl/Dockerfile' },
  { name: 'grafana/agent-operator', make: 'make operator-image', path: 'cmd/grafana-agent-operator/Dockerfile' },
];

(
  std.map(function(container) pipelines.linux('Check Linux container (%s)' % container.name) {
    trigger: {
      event: ['pull_request'],
      paths: [container.path],
    },
    steps: [{
      name: 'Build container',
      image: build_image.linux,
      volumes: [{
        name: 'docker',
        path: '/var/run/docker.sock',
      }],
      commands: [container.make],
    }],
    volumes: [{
      name: 'docker',
      host: {
        path: '/var/run/docker.sock',
      },
    }],
  }, linux_containers)
) + [
  pipelines.windows('Check Windows containers') {
    trigger: {
      event: ['pull_request'],
      paths: ['**/Dockerfile.windows'],
    },
    steps: [{
      name: 'Build container',
      image: build_image.windows,
      volumes: [{
        name: 'docker',
        path: '//./pipe/docker_engine/',
      }],
      commands: [
        'git config --global --add safe.directory C:/drone/src/',
        '& "C:/Program Files/git/bin/bash.exe" -c ./tools/ci/docker-containers-windows',
      ],
    }],
    volumes: [{
      name: 'docker',
      host: {
        path: '//./pipe/docker_engine/',
      },
    }],
  },
]
