local new_secret(name) = {
  kind: 'secret',
  name: name,

  getFrom(path, name):: self {
    get: { path: path, name: name },
  },
};

[
  new_secret('dockerconfigjson').getFrom(path='secret/data/common/gcr', name='.dockerconfigjson'),
  new_secret('gcr_admin').getFrom(path='infra/data/ci/gcr-admin', name='.dockerconfigjson'),
  new_secret('gh_token').getFrom(path='infra/data/ci/github/grafanabot', name='pat'),
  new_secret('gpg_public_key').getFrom(path='infra/data/ci/packages-publish/gpg', name='public-key'),
  new_secret('gpg_private_key').getFrom(path='infra/data/ci/packages-publish/gpg', name='private-key'),
  new_secret('gpg_passphrase').getFrom(path='infra/data/ci/packages-publish/gpg', name='passphrase'),
]
