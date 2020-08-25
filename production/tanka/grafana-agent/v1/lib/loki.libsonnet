local scrape_k8s_logs = import '../internal/kubernetes_logs.libsonnet';

{
  // withLokiConfig adds a Loki config to collect logs.
  //
  // For the full list of options, refer to the configuration reference:
  // https://github.com/grafana/agent/blob/master/docs/configuration-reference.md#loki_config
  withLokiConfig(config):: {
    assert std.objectHasAll(self, '_mode') : |||
      withLokiConfig must be merged with the result of calling new.
    |||,
    _loki_config:: config,
  },

  // newLokiClient creates a new client object. Results from this can be passed into 
  // withLokiClients.
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
  newLokiClient(client_config):: 
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

  // withLokiClients adds clients to send logs to. At least one client must be
  // present. Clients can be created by calling newLokiClient or by creating 
  // an object that conforms to the Promtail client_config schema specified 
  // here:
  //
  // https://grafana.com/docs/loki/latest/clients/promtail/configuration/#client_config
  // 
  // withLokiClients should be merged with the result of withLokiConfig.
  withLokiClients(clients):: {
    assert std.objectHasAll(self, '_loki_config') : |||
      withLokiClients must be merged with the result of calling withLokiConfig.
    |||,

    _loki_config+:: {
      clients: if std.isArray(clients) then clients else [clients],
    },
  },

  // scrapeKubernetesLogs defines a Loki config that can collect logs from
  // Kubernetes pods.
  scrapeKubernetesLogs: scrape_k8s_logs.newKubernetesLogsCollector(),
}
