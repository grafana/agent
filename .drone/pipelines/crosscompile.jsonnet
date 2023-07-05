local build_image = import '../util/build_image.jsonnet';
local pipelines = import '../util/pipelines.jsonnet';

local go_tags = {
  linux: 'builtinassets promtail_journal_enabled',
  windows: 'builtinassets',
  darwin: 'builtinassets',
  freebsd: 'builtinassets',
};

local os_arch_tuples = [
  // Linux
  { name: 'Linux amd64', os: 'linux', arch: 'amd64' },
  { name: 'Linux arm64', os: 'linux', arch: 'arm64' },
  { name: 'Linux ppc64le', os: 'linux', arch: 'ppc64le' },
  { name: 'Linux s390x', os: 'linux', arch: 's390x' },

  // Darwin
  { name: 'macOS Intel', os: 'darwin', arch: 'amd64' },
  { name: 'macOS Apple Silicon', os: 'darwin', arch: 'arm64' },

  // Windows
  { name: 'Windows amd64', os: 'windows', arch: 'amd64' },

  // FreeBSD
  { name: 'FreeBSD amd64', os: 'freebsd', arch: 'amd64' },
];


local targets = [
  'agent',
  'agent-flow',
  'agentctl',
  'operator',
];

local targets_boringcrypto = [
  'agent-boringcrypto',
];

local os_arch_types_boringcrypto = [
  // Linux boringcrypto
  { name: 'Linux amd64 boringcrypto', os: 'linux', arch: 'amd64', experiment: 'boringcrypto' },
  { name: 'Linux arm64 boringcrypto', os: 'linux', arch: 'arm64', experiment: 'boringcrypto' },
];


std.flatMap(function(target) (
  std.map(function(platform) (
    pipelines.linux('Build %s (%s)' % [target, platform.name]) {
      local env = {
        GOOS: platform.os,
        GOARCH: platform.arch,
        GOARM: if 'arm' in platform then platform.arm else '',

        target: target,

        tags: go_tags[platform.os],
      },

      trigger: {
        event: ['pull_request'],
      },
      steps: [{
        name: 'Build',
        image: build_image.linux,
        commands: [
          'make generate-ui',
          'GO_TAGS="%(tags)s" GOOS=%(GOOS)s GOARCH=%(GOARCH)s GOARM=%(GOARM)s make %(target)s' % env,
        ],
      }],
    }
  ), os_arch_tuples)
), targets) +
std.flatMap(function(target) (
  std.map(function(platform) (
    pipelines.linux('Build %s (%s)' % [target, platform.name]) {
      local env = {
        GOOS: platform.os,
        GOARCH: platform.arch,
        GOARM: if 'arm' in platform then platform.arm else '',
        GOEXPERIMENT: platform.experiment,

        target: target,

        tags: go_tags[platform.os],
      },

      trigger: {
        event: ['pull_request'],
      },
      steps: [{
        name: 'Build',
        image: build_image.linux,
        commands: [
          'make generate-ui',
          'GO_TAGS="%(tags)s" GOOS=%(GOOS)s GOARCH=%(GOARCH)s GOARM=%(GOARM)s GOEXPERIMENT=%(GOEXPERIMENT)s make %(target)s' % env,
        ],
      }],
    }
  ), os_arch_types_boringcrypto)
), targets_boringcrypto)
