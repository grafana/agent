local build_image_version = std.extVar('__build-image-version');

local condition(verb) = {
  tagMaster: {
    ref: {
      [verb]: [
        'refs/heads/master',
        'refs/tags/v*',
      ],
    },
  },
};

local pipeline(name) = {
  kind: 'pipeline',
  name: name,
  steps: [],
};

local run(name, commands) = {
  name: name,
  // TODO: grafana/agent-build-image?
  image: 'grafana/loki-build-image:%s' % build_image_version,
  commands: commands,
};

local make(target, container=true) = run(target, [
  'make ' + (if !container then 'BUILD_IN_CONTAINER=false ' else '') + target,
]);

[
  pipeline('check') {
    workspace: {
      base: '/src',
      path: 'agent',
    },
    steps: [
      make('test', container=false) { depends_on: ['clone'] },
      make('lint', container=false) { depends_on: ['clone'] },
    ],
  },
]
