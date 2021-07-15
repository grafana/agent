local optionals = import 'ext/optionals.libsonnet';
local secrets = import 'ext/secrets.libsonnet';

local new_safe_tls_config = import './safe_tls_config.libsonnet';

// @param {string} namespace
// @param {TLSConfig} config
function(namespace, config) new_safe_tls_config(namespace, config.SafeTLSConfig) + {
  // Local configurations for ca_file, cert_file, and key_file take precedence
  // over the SafeTLSConfig. Check local settings first and then fall back
  // to the safe setings.

  ca_file:
    local unsafe = optionals.string(config.CAFile);
    if unsafe == null then super.ca_file else unsafe,

  cert_file:
    local unsafe = optionals.string(config.CertFile);
    if unsafe == null then super.cert_file else unsafe,

  key_file:
    local unsafe = optionals.string(config.KeyFile);
    if unsafe == null then super.key_file else unsafe,
}
