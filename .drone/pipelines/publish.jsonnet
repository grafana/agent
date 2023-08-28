local build_image = import '../util/build_image.jsonnet';
local pipelines = import '../util/pipelines.jsonnet';
local secrets = import '../util/secrets.jsonnet';
local ghTokenFilename = '/drone/src/gh-token.txt';

// job_names gets the list of job names for use in depends_on.
local job_names = function(jobs) std.map(function(job) job.name, jobs);

local linux_containers = ['agent','agent-boringcrypto', 'agentctl', 'agent-operator', 'smoke', 'crow'];
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
        DOCKER_LOGIN: secrets.docker_login.fromSecret,
        DOCKER_PASSWORD: secrets.docker_password.fromSecret,
        GCR_CREDS: secrets.gcr_admin.fromSecret,
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
        DOCKER_LOGIN: secrets.docker_login.fromSecret,
        DOCKER_PASSWORD: secrets.docker_password.fromSecret,
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
              "git_committer_name": "updater-for-ci[bot]",
              "git_author_name": "updater-for-ci[bot]",
              "git_committer_email": "119986603+updater-for-ci[bot]@users.noreply.github.com",
              "git_author_email": "119986603+updater-for-ci[bot]@users.noreply.github.com",
              "destination_branch": "master",
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
                },
                {
                  "file_path": "ksonnet/environments/pyroscope-ebpf/waves/ebpf.libsonnet",
                  "jsonnet_key": "dev_canary",
                  "jsonnet_value_file": ".image-tag"
                }
              ]
            }
          |||,
          github_app_id: secrets.updater_app_id.fromSecret,
          github_app_installation_id: secrets.updater_app_installation_id.fromSecret,
          github_app_private_key: secrets.updater_private_key.fromSecret,
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
    image_pull_secrets: ['dockerconfigjson'],
    steps: [
      {
        name: 'Generate GitHub token',
        image: 'us.gcr.io/kubernetes-dev/github-app-secret-writer:latest',
        environment: {
          GITHUB_APP_ID: secrets.updater_app_id.fromSecret,
          GITHUB_APP_INSTALLATION_ID: secrets.updater_app_installation_id.fromSecret,
          GITHUB_APP_PRIVATE_KEY: secrets.updater_private_key.fromSecret,
        },
        commands: [
          '/usr/bin/github-app-external-token > %s' % ghTokenFilename
        ]
      },
      {
        name: 'Publish release',
        image: build_image.linux,
        volumes: [{
          name: 'docker',
          path: '/var/run/docker.sock',
        }],
        environment: {
          DOCKER_LOGIN: secrets.docker_login.fromSecret,
          DOCKER_PASSWORD: secrets.docker_password.fromSecret,
          GPG_PRIVATE_KEY: secrets.gpg_private_key.fromSecret,
          GPG_PUBLIC_KEY: secrets.gpg_public_key.fromSecret,
          GPG_PASSPHRASE: secrets.gpg_passphrase.fromSecret,
        },
        commands: [
          'export GITHUB_TOKEN=$(cat %s)' % ghTokenFilename,
          'docker login -u $DOCKER_LOGIN -p $DOCKER_PASSWORD',
          'make -j4 RELEASE_BUILD=1 VERSION=${DRONE_TAG} dist',
          |||
            VERSION=${DRONE_TAG} RELEASE_DOC_TAG=$(echo ${DRONE_TAG} | awk -F '.' '{print $1"."$2}') ./tools/release
          |||,
        ],
      }
    ],
    volumes: [{
      name: 'docker',
      host: { path: '/var/run/docker.sock' },
    }],
  },
]
