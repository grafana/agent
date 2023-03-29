local build_image = import '../util/build_image.jsonnet';
local pipelines = import '../util/pipelines.jsonnet';

[
  pipelines.linux('Test Linux system packages') {
    trigger: {
      event: ['pull_request'],
      paths: [
        'packaging/**',
        'Makefile',
      ],
    },
    steps: [{
      name: 'Test Linux system packages',
      image: build_image.linux,
      volumes: [{
        name: 'docker',
        path: '/var/run/docker.sock',
      }],
      commands: [
        'DOCKER_OPTS="" make dist/grafana-agent-linux-amd64',
        'DOCKER_OPTS="" make dist/grafana-agentctl-linux-amd64',
        'DOCKER_OPTS="" make dist.temp/grafana-agent-flow-linux-amd64',
        'DOCKER_OPTS="" make test-packages',
      ],
    }],
    volumes: [{
      name: 'docker',
      host: {
        path: '/var/run/docker.sock',
      },
    }],
  },
]
