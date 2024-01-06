---
aliases:
- ../../../../configuration/integrations/integrations-next/eventhandler-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/integrations/integrations-next/eventhandler-config/
- /docs/grafana-cloud/send-data/agent/static/configuration/integrations/integrations-next/eventhandler-config/
canonical: https://grafana.com/docs/agent/latest/static/configuration/integrations/integrations-next/eventhandler-config/
description: Learn about eventhandler_config next
title: eventhandler_config next
---

# eventhandler_config next

`eventhandler_config` configures the Kubernetes eventhandler integration. This
integration watches
[Event](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#event-v1-core)
resources in a Kubernetes cluster and forwards them as log entries to a Loki
sink. This integration depends on the experimental `integrations-next` feature
being enabled.

On restart, the integration will look for a cache file (configured using
`cache_path`) that stores the last shipped event. This file is optional, and if
present, will be used to avoid double-shipping events if Agent or the
integration restarts. Kubernetes expires events after 60 minutes, so events
older than 60 minutes ago will never be shipped.

To use the cache feature and maintain state in a Kubernetes environment, a
[StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
must be used. Sample manifests are provided at the bottom of this doc. Please
adjust these according to your deployment preferences. You can also use a
Deployment, however the presence of the cache file will not be guaranteed and
the integration may ship duplicate entries in the event of a restart. Loki does
not yet support entry deduplication for the A->B->A case, so further
deduplication can only take place at the Grafana / front-end layer (Grafana
Explore does provide some deduplication features for Loki datasources).

This integration uses Grafana Agent's embedded Loki-compatible `logs` subsystem
to ship entries, and a logs client and sink must be configured to use the
integration. Please see the sample Agent config below for an example
configuration.
[Pipelines](/docs/loki/latest/clients/promtail/pipelines/)
and relabel configuration are not yet supported, but these features will be
added soon. You should use the `job=eventhandler cluster=...` labels to query
your events (you can then use LogQL on top of the result set).

If not running the integration in-cluster, the integration will use
`kubeconfig_path` to search for a valid Kubeconfig file, defaulting to a
kubeconfig in the user's home directory. If running in-cluster, the appropriate
`ServiceAccount` and Roles must be defined. Sample manifests are provided
below.

Configuration reference:

```yaml
  # Provide an explicit value to uniquely identify this instance of the
  # integration. If not provided, a reasonable default will be inferred based
  # on the integration.
  #
  # The value here must be unique across all instances of the same integration.
  [instance: <string>]

  ## Eventhandler hands watched events off to promtail using a promtail
  ## client channel. This parameter configures how long to wait (in seconds) on the channel
  ## before abandoning and moving on.
  [send_timeout: <int> | default = 60]

  ## Configures the path to a kubeconfig file. If not set, will fall back to using
  ## an in-cluster config. If this fails, will fall back to checking the user's home
  ## directory for a kubeconfig.
  [kubeconfig_path: <string>]

  ## Path to a cache file that will store the last timestamp for a shipped event and events
  ## shipped for that timestamp. Used to prevent double-shipping on integration restart.
  [cache_path: <string> | default = "./.eventcache/eventhandler.cache"]

  ## Name of logs subsystem instance to hand log entries off to.
  [logs_instance: <string> | default = "default"]

  ## K8s informer resync interval (seconds). You should use defaults here unless you are
  ## familiar with K8s informers.
  [informer_resync: <int> | default = 120]

  ## The integration will flush the last event shipped out to disk every flush_interval seconds.
  [flush_interval: <int> | default = 10]

  ## If you would like to limit events to a given namespace, use this parameter.
  [namespace: <string>]

  ## Configure extra labels to add to log lines
  extra_labels:
    { <string>: <string> }

  ## Format of the log line. The possible values are "logfmt" and "json".
  ## The values are also LogQL parsers, which can be used for processing the logs
  [log_format: <string> | default = "logfmt"]
```

Sample agent config:

```yaml
server:
  log_level: info

integrations:
  eventhandler:
    cache_path: "/etc/eventhandler/eventhandler.cache"

logs:
  configs:
  - name: default
    clients:
    - url: https://logs-prod-us-central1.grafana.net/api/prom/push
      basic_auth:
        username: YOUR_LOKI_USER
        password: YOUR_LOKI_API_KEY
      external_labels:
        cluster: "cloud"
    positions:
      filename: /tmp/positions0.yaml
```

Be sure to replace the Loki credentials with the appropriate values.

Sample StatefulSet manifests. Please adjust these according to your needs:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: grafana-agent-eventhandler
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: grafana-agent-eventhandler
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: grafana-agent-eventhandler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: grafana-agent-eventhandler
subjects:
- kind: ServiceAccount
  name: grafana-agent-eventhandler
  namespace: default
---
apiVersion: v1
kind: Service
metadata:
  name: grafana-agent-eventhandler-svc
spec:
  ports:
  - port: 12345
    name: http-metrics
  clusterIP: None
  selector:
    name: grafana-agent-eventhandler
---
kind: ConfigMap
metadata:
  name: grafana-agent-eventhandler
  namespace: default
apiVersion: v1
data:
  agent.yaml: |
    server:
      log_level: info

    integrations:
      eventhandler:
        cache_path: "/etc/eventhandler/eventhandler.cache"

    logs:
      configs:
      - name: default
        clients:
        - url: https://logs-prod-us-central1.grafana.net/api/prom/push
          basic_auth:
            username: YOUR_LOKI_USER
            password: YOUR_LOKI_API_KEY
          external_labels:
            cluster: "cloud"
        positions:
          filename: /tmp/positions0.yaml
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: grafana-agent-eventhandler
  namespace: default
spec:
  serviceName: "grafana-agent-eventhandler-svc"
  selector:
    matchLabels:
      name: grafana-agent-eventhandler
  replicas: 1
  template:
    metadata:
      labels:
        name: grafana-agent-eventhandler
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: agent
        image: grafana/agent:main
        imagePullPolicy: IfNotPresent
        args:
        - -config.file=/etc/agent/agent.yaml
        - -enable-features=integrations-next
        - -server.http.address=0.0.0.0:12345
        command:
        - /bin/grafana-agent
        env:
        - name: HOSTNAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        ports:
        - containerPort: 12345
          name: http-metrics
        volumeMounts:
        - name: grafana-agent
          mountPath: /etc/agent
        - name: eventhandler-cache
          mountPath: /etc/eventhandler
      serviceAccount: grafana-agent-eventhandler
      volumes:
        - configMap:
            name: grafana-agent-eventhandler
          name: grafana-agent
  volumeClaimTemplates:
  - metadata:
      name: eventhandler-cache
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
```
