local scrape_k8s_logs = import '../internal/kubernetes_logs.libsonnet';
local k = import 'ksonnet-util/kausal.libsonnet';

local container = k.core.v1.container;

{
  // withLogsConfig adds a Logs config to collect logs.
  //
  // For the full list of options, refer to the configuration reference:
  // https://grafana.com/docs/agent/latest/configuration/logs-config/
  withLogsConfig(config):: {
    assert std.objectHasAll(self, '_mode') : |||
      withLogsConfig must be merged with the result of calling new.
    |||,
    _logs_config:: config,
  },

  // newLogsClient creates a new client object. Results from this can be passed into
  // withLogsClients.
  //
  // client_config should be an object of the following shape:
  //
  // {
  //   scheme: 'https', // or http
  //   hostname: 'logs-us-central1.grafana.net', // replace with hostname to use
  //   username: '', // OPTIONAL username for Loki API connection
  //   password: '', // OPTIONAL password for Loki API connection
  //   external_labels: {}, // OPTIONAL labels to set for connection
  // }
  newLogsClient(client_config)::
    {
      url: (
        if std.objectHasAll(client_config, 'username') then
          '%(scheme)s://%(username)s:%(password)s@%(hostname)s/loki/api/v1/push' % client_config
        else
          '%(scheme)s://%(hostname)s/loki/api/v1/push' % client_config
      ),
    } + (
      if std.objectHasAll(client_config, 'external_labels')
      then { external_labels: client_config.external_labels }
      else {}
    ),

  // withLogsClients adds clients to send logs to. At least one client must be
  // present. Clients can be created by calling newLogsClient or by creating
  // an object that conforms to the Promtail client_config schema specified
  // here:
  //
  // https://grafana.com/docs/loki/latest/clients/promtail/configuration/#client_config
  //
  // withLogsClients should be merged with the result of withLogsConfig.
  withLogsClients(clients):: {
    assert std.objectHasAll(self, '_logs_config') : |||
      withLogsClients must be merged with the result of calling withLogsConfig.
    |||,

    _logs_config+:: {
      clients: if std.isArray(clients) then clients else [clients],
    },
  },

  // logsPermissionsMixin mutates the container and deployment to work with
  // reading Docker container logs.
  logsPermissionsMixin:: {
    container+::
      container.mixin.securityContext.withPrivileged(true) +
      container.mixin.securityContext.withRunAsUser(0),

    agent+:
      // For reading docker containers. /var/log is used for the positions file
      // and shouldn't be set to readonly.
      k.util.hostVolumeMount('varlog', '/var/log', '/var/log') +
      k.util.hostVolumeMount('varlibdockercontainers', '/var/lib/docker/containers', '/var/lib/docker/containers', readOnly=true) +

      // For reading journald
      k.util.hostVolumeMount('etcmachineid', '/etc/machine-id', '/etc/machine-id', readOnly=true),
  },

  // scrapeKubernetesLogs defines a Logs config that can collect logs from
  // Kubernetes pods.
  scrapeKubernetesLogs: scrape_k8s_logs.newKubernetesLogsCollector(),
}
