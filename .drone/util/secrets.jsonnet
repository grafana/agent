local newSecret(name) = {
  kind: 'secret',
  name: name,

  getFrom(path, name):: self {
    get: { path: path, name: name },
  },

  fromSecret:: local secret = self; { from_secret: secret.name },
};

{
  dockerconfigjson: newSecret('dockerconfigjson').getFrom(path='secret/data/common/gcr', name='.dockerconfigjson'),
  gcr_admin: newSecret('gcr_admin').getFrom(path='infra/data/ci/gcr-admin', name='.dockerconfigjson'),
  gh_token: newSecret('gh_token').getFrom(path='infra/data/ci/github/grafanabot', name='pat'),
  private_key: newSecret('private_key').getFrom(path='infra/data/ci/agent/githubapp', name='private-key'),
  app_id: newSecret('app_id').getFrom(path='infra/data/ci/agent/githubapp', name='app-id'),
  app_installation_id: newSecret('app_installation_id').getFrom(path='infra/data/ci/agent/githubapp', name='app-installation-id'),
  gpg_public_key: newSecret('gpg_public_key').getFrom(path='infra/data/ci/packages-publish/gpg', name='public-key'),
  gpg_private_key: newSecret('gpg_private_key').getFrom(path='infra/data/ci/packages-publish/gpg', name='private-key'),
  gpg_passphrase: newSecret('gpg_passphrase').getFrom(path='infra/data/ci/packages-publish/gpg', name='passphrase'),
  docker_login: newSecret('docker_login').getFrom(path='infra/data/ci/docker_hub', name='username'),
  docker_password: newSecret('docker_password').getFrom(path='infra/data/ci/docker_hub', name='password'),

  asList:: [self[k] for k in std.sort(std.objectFields(self))],
}
