local build_image = import '../util/build_image.jsonnet';
local pipelines = import '../util/pipelines.jsonnet';

// job_names gets the list of job names for use in depends_on.
local job_names = function(jobs) std.map(function(job) job.name, jobs);

local linux_containers = ['agent', 'agentctl', 'agent-operator', 'smoke', 'crow'];
local linux_containers_jobs = std.map(function(container) (
  pipelines.linux('Publish Linux %s container' % container) {
    trigger: {
      ref: [
        'refs/heads/main',
        'refs/tags/v*',
      ],
    },
    steps: [{
      // We only need to run this once per machine, so it's OK if it fails. It
      // is also likely to fail when run in parallel on the same machine.
      name: 'Configure QEMU',
      image: build_image.linux,
      failure: 'ignore',
      volumes: [{
        name: 'docker',
        path: '/var/run/docker.sock',
      }],
      commands: [
        'docker run --rm --privileged multiarch/qemu-user-static --reset -p yes',
      ],
    }, {
      name: 'Publish container',
      image: build_image.linux,
      volumes: [{
        name: 'docker',
        path: '/var/run/docker.sock',
      }],
      environment: {
        DOCKER_LOGIN: { from_secret: 'DOCKER_LOGIN' },
        DOCKER_PASSWORD: { from_secret: 'DOCKER_PASSWORD' },
        GCR_CREDS: { from_secret: 'gcr_admin' },
      },
      commands: [
        'mkdir -p $HOME/.docker',
        'printenv GCR_CREDS > $HOME/.docker/config.json',
        'docker login -u $DOCKER_LOGIN -p $DOCKER_PASSWORD',

        // Create a buildx worker for our cross platform builds.
        'docker buildx create --name multiarch-agent-%s-${DRONE_COMMIT_SHA} --driver docker-container --use' % container,

        './tools/ci/docker-containers %s' % container,

        'docker buildx rm multiarch-agent-%s-${DRONE_COMMIT_SHA}' % container,
      ],
    }],
    volumes: [{
      name: 'docker',
      host: { path: '/var/run/docker.sock' },
    }],
  }
), linux_containers);

local windows_containers = ['agent', 'agentctl'];
local windows_containers_jobs = std.map(function(container) (
  pipelines.windows('Publish Windows %s container' % container) {
    trigger: {
      ref: [
        'refs/heads/main',
        'refs/tags/v*',
      ],
    },
    steps: [{
      name: 'Build containers',
      image: build_image.windows,
      volumes: [{
        name: 'docker',
        path: '//./pipe/docker_engine/',
      }],
      environment: {
        DOCKER_LOGIN: { from_secret: 'DOCKER_LOGIN' },
        DOCKER_PASSWORD: { from_secret: 'DOCKER_PASSWORD' },
      },
      commands: [
        '& "C:/Program Files/git/bin/bash.exe" ./tools/ci/docker-containers-windows %s' % container,
      ],
    }],
    volumes: [{
      name: 'docker',
      host: { path: '//./pipe/docker_engine/' },
    }],
  }
), windows_containers);

linux_containers_jobs + windows_containers_jobs + [
  pipelines.linux('Deploy to deployment_tools') {
    trigger: {
      ref: ['refs/heads/main'],
    },
    image_pull_secrets: ['dockerconfigjson'],
    steps: [
      {
        name: 'Create .image-tag',
        image: 'alpine',
        commands: [
          'apk update && apk add git',
          'echo "$(sh ./tools/image-tag)" > .tag-only',
          'echo "grafana/agent:$(sh ./tools/image-tag)" > .image-tag',
        ],
      },
      {
        name: 'Update deployment_tools',
        image: 'us.gcr.io/kubernetes-dev/drone/plugins/updater',
        settings: {
          config_json: |||
            {
              "destination_branch": "master",
              "pull_request_branch_prefix": "cd-agent",
              "pull_request_enabled": false,
              "pull_request_team_reviewers": [
                "agent-squad"
              ],
              "repo_name": "deployment_tools",
              "update_jsonnet_attribute_configs": [
                {
                  "file_path": "ksonnet/environments/kowalski/dev-us-central-0.kowalski-dev/main.jsonnet",
                  "jsonnet_key": "agent_image",
                  "jsonnet_value_file": ".image-tag"
                },
                {
                  "file_path": "ksonnet/environments/grafana-agent/waves/agent.libsonnet",
                  "jsonnet_key": "dev_canary",
                  "jsonnet_value_file": ".image-tag"
                }
              ]
            }
          |||,
          github_token: {
            from_secret: 'gh_token',
          },
        },
      },
    ],
    depends_on: job_names(linux_containers_jobs),
  },

  pipelines.linux('Publish release') {
    trigger: {
      ref: ['refs/tags/v*'],
    },
    depends_on: job_names(linux_containers_jobs + windows_containers_jobs),
    steps: [{
      name: 'Publish release',
      image: build_image.linux,
      volumes: [{
        name: 'docker',
        path: '/var/run/docker.sock',
      }],
      environment: {
        DOCKER_LOGIN: { from_secret: 'DOCKER_LOGIN' },
        DOCKER_PASSWORD: { from_secret: 'DOCKER_PASSWORD' },
        GITHUB_TOKEN: { from_secret: 'GITHUB_KEY' },
        GPG_PRIVATE_KEY: { from_secret: 'gpg_private_key' },
        GPG_PUBLIC_KEY: { from_secret: 'gpg_public_key' },
        GPG_PASSPHRASE: { from_secret: 'gpg_passphrase' },
      },
      commands: [
        'docker login -u $DOCKER_LOGIN -p $DOCKER_PASSWORD',
        'make -j4 RELEASE_BUILD=1 VERSION=${DRONE_TAG} dist',
        |||
          VERSION=${DRONE_TAG} RELEASE_DOC_TAG=$(echo ${DRONE_TAG} | awk -F '.' '{print $1"."$2}') ./tools/release
        |||,
      ],
    }],
    volumes: [{
      name: 'docker',
      host: { path: '/var/run/docker.sock' },
    }],
  },
]
