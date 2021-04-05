local optionals = import '../ext/optionals.libsonnet';
local secrets = import '../ext/secrets.libsonnet';

local new_tls_config = import './tls_config.libsonnet';

// kubernetes_sd_config returns a kubernetes_sd_config entry.
//
// Params:
//    namespace: namespace of the GrafanaAgent resource.
//    namespaces: list of namespaces to discover resources in.
//    apiServer: APIServerConfig to use for discovery.
//    role: role of resources to discover.
function(
  namespace,
  namespaces,
  apiServer,
  role,
) {
  role: role,
  namespaces: if std.length(namespaces) > 0 then {
    names: namespaces,
  },

  apiServer: if apiServer != null then {
    basic_auth: if apiServer.BasicAuth != null then {
      username: secrets.valueForSecret(namespace, apiServer.BasicAuth.Username),
      password: secrets.valueForSecret(namespace, apiServer.BasicAuth.Password),
    },
    bearer_token: optionals.string(apiServer.BearerToken),
    bearer_token_file: optionals.string(apiServer.BearerTokenFile),
    tls_config:
      if apiServer.TLSConfig != null then new_tls_config(namespace, apiServer.TLSConfig),
  },
}
