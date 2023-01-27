{
  linux(name):: {
    kind: 'pipeline',
    type: 'docker',
    name: name,
    platform: {
      os: 'linux',
      arch: 'amd64',
    },
  },

  windows(name):: {
    kind: 'pipeline',
    type: 'docker',
    name: name,
    platform: {
      arch: 'amd64',
      os: 'windows',
      version: '1809',
    },
  },
}
