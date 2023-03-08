local build_image = import '../util/build_image.jsonnet';
local pipelines = import '../util/pipelines.jsonnet';

[
  pipelines.linux('Lint') {
    trigger: {
      event: ['pull_request'],
    },
    steps: [{
      name: 'Lint',
      image: build_image.linux,
      commands: [
        'apt-get update -y && apt-get install -y libsystemd-dev',
        'make lint',
      ],
    }],
  },

  pipelines.linux('Test dashboards') {
    trigger: {
      event: ['pull_request'],
    },
    steps: [{
      name: 'Regenerate dashboards',
      image: build_image.linux,

      commands: [
        'make generate-dashboards',
        'ERR_MSG="Dashboard definitions are out of date. Please run \'make generate-dashboards\' and commit changes!"',
        // "git status --porcelain" reports if there's any new, modified, or deleted files.
        'if [ ! -z "$(git status --porcelain)" ]; then echo $ERR_MSG >&2; exit 1; fi',
      ],
    }],
  },

  pipelines.linux('Test manifests') {
    trigger: {
      event: ['pull_request'],
    },
    steps: [{
      name: 'Regenerate environment manifests',
      image: build_image.linux,

      commands: [
        'make generate-manifests',
        'ERR_MSG="The environment manifests are out of date. Please run \'make generate-manifests\' and commit changes!"',
        // "git status --porcelain" reports if there's any new, modified, or deleted files.
        'if [ ! -z "$(git status --porcelain)" ]; then echo $ERR_MSG >&2; exit 1; fi',
      ],
    }],
  },

  pipelines.linux('Test') {
    trigger: {
      event: ['pull_request'],
    },
    steps: [{
      name: 'Run Go tests',
      image: build_image.linux,
      volumes: [{
        name: 'docker',
        path: '/var/run/docker.sock',
      }],

      commands: [
        // The operator tests require K8S_USE_DOCKER_NETWORK=1 to be set when
        // tests are being run inside of a Docker container so it can access the
        // created k3d cluster properly.
        'K8S_USE_DOCKER_NETWORK=1 make test',
      ],
    }],
    volumes: [{
      name: 'docker',
      host: {
        path: '/var/run/docker.sock',
      },
    }],
  },

  pipelines.windows('Test (Windows)') {
    trigger: {
      event: ['pull_request'],
    },
    steps: [{
      name: 'Run Go tests',
      image: build_image.windows,
      environment: {
        ASSUME_NO_MOVING_GC_UNSAFE_RISK_IT_WITH: 'go1.20',
      },
      commands: ['go test -tags="nodocker,nonetwork" ./...'],
    }],
  },
]
