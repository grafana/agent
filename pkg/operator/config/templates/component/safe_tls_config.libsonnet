local optionals = import '../ext/optionals.libsonnet';
local secrets = import '../ext/secrets.libsonnet';

// config is expected to be a SafeTLSConfig.
function(namespace, config) {
  ca_file: secrets.pathForSelector(namespace, config.CA),
  cert_file: secrets.pathForSelector(namespace, config.Cert),
  key_file: secrets.pathForSecret(namespace, config.KeySecret),

  server_name: optionals.string(config.ServerName),
  insecure_skip_verify: optionals.bool(config.InsecureSkipVerify),
}
