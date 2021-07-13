local marshal = import 'ext/marshal.libsonnet';
local optionals = import 'ext/optionals.libsonnet';
local secrets = import 'ext/secrets.libsonnet';
local k8s = import 'utils/k8s.libsonnet';

local new_client = import 'component/logs/client.libsonnet';

// Generates a logs_instance.
//
// @param {string} agentNamespace - namespace of the GrafanaAgent
// @param {LogsSubsystemSpec} global - global logs settings & defaults
// @param {LogsInstance} instance
// @param {APIServerConfig} apiServer
function(agentNamespace, global, instance, apiServer) {
  local meta = instance.Instance.ObjectMeta,
  local spec = instance.Instance.Spec,

  name: '%s/%s' % [meta.Namespace, meta.Name],

  // Figure out what set of clients to use and what namespace they're in.
  // We'll only use the global set of clients if the local LogsInstance doesn't
  // have a set of clients defined.
  //
  // Local clients come from the namespace of the LogsInstance and global
  // clients from the Agent's namespace.
  local clients =
    if std.length(spec.Clients) != 0
    then { ns: meta.Namespace, list: spec.Clients }
    else { ns: agentNamespace, list: global.Clients },

  clients: optionals.array(std.map(
    function(spec) new_client(clients.ns, spec),
    clients.list,
  )),

  scrape_configs: optionals.array([]),
}
