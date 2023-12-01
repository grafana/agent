---
aliases:
- /docs/grafana-cloud/monitor-infrastructure/agent/static/set-up/deploy-agent/
- /docs/grafana-cloud/send-data/agent/static/set-up/deploy-agent/
canonical: https://grafana.com/docs/agent/latest/static/set-up/deploy-agent/
description: Learn how to deploy Grafana Agent in different topologies
menuTitle: Deploy static mode
title: Deploy Grafana Agent in static mode
weight: 300
---

{{< docs/shared source="agent" lookup="/deploy-agent.md" version="<AGENT_VERSION>" >}}

## For scalable ingestion of traces

For small workloads, it is normal to have just one Agent handle all incoming
spans with no need of load balancing. However, for large workloads there it
is desirable to spread out the load of ingesting spans over multiple Agent
instances.

To scale the Agent for trace ingestion, do the following:
1. Set up the `load_balancing` section of the Agent's `traces` config.
2. Start multiple Agent instances, all with the same configuration, so that:
   * Each Agent load balances using the same strategy.
   * Each Agent processes spans in the same way.
3. The cluster of Agents is now setup for load balancing. It works as follows:
   1. Any of the Agents can receive spans from instrumented applications via the configured `receivers`.
   2. When an Agent firstly receives spans, it will forward them to any of the Agents in the cluster according to the `load_balancing` configuration.

<!-- For more detail, send people over to the load_balancing section in traces_config -->

### tail_sampling
Agents configured with `tail_sampling` must have all spans for 
a given trace in order to work correctly. If some of the spans for a trace end up
in a different Agent, `tail_sampling` will not sample correctly.

### spanmetrics
<!-- TODO: Also talk about span metrics -->

### service_graphs
<!-- TODO: Also talk about service_graphs -->

### Example Kubernetes configuration
{{< collapse title="Example Kubernetes configuration with DNS load balancing" >}}
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: grafana-cloud-monitoring
---
apiVersion: v1
kind: Service
metadata:
  name: agent-traces
  namespace: grafana-cloud-monitoring
spec:
  ports:
  - name: agent-traces-otlp-grpc
    port: 9411
    protocol: TCP
    targetPort: 9411
  selector:
    name: agent-traces
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k6-trace-generator
  namespace: grafana-cloud-monitoring
spec:
  minReadySeconds: 10
  replicas: 1
  revisionHistoryLimit: 1
  selector:
    matchLabels:
      name: k6-trace-generator
  template:
    metadata:
      labels:
        name: k6-trace-generator
    spec:
      containers:
      - env:
        - name: ENDPOINT
          value: agent-traces-headless.grafana-cloud-monitoring.svc.cluster.local:9411
        image: ghcr.io/grafana/xk6-client-tracing:v0.0.2
        imagePullPolicy: IfNotPresent
        name: k6-trace-generator
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-traces
  namespace: grafana-cloud-monitoring
spec:
  minReadySeconds: 10
  replicas: 3
  revisionHistoryLimit: 1
  selector:
    matchLabels:
      name: agent-traces
  template:
    metadata:
      labels:
        name: agent-traces
    spec:
      containers:
      - args:
        - -config.file=/etc/agent/agent.yaml
        command:
        - /bin/grafana-agent
        image: grafana/agent:v0.38.0
        imagePullPolicy: IfNotPresent
        name: agent-traces
        ports:
        - containerPort: 9411
          name: otlp-grpc
          protocol: TCP
        - containerPort: 34621
          name: agent-lb
          protocol: TCP
        volumeMounts:
        - mountPath: /etc/agent
          name: agent-traces
      volumes:
      - configMap:
          name: agent-traces
        name: agent-traces
---
apiVersion: v1
kind: Service
metadata:
  name: agent-traces-headless
  namespace: grafana-cloud-monitoring
spec:
  clusterIP: None
  ports:
  - name: agent-lb
    port: 34621
    protocol: TCP
    targetPort: agent-lb
  selector:
    name: agent-traces
  type: ClusterIP
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: agent-traces
  namespace: grafana-cloud-monitoring
data:
  agent.yaml: |
    traces:
      configs:
      - name: default
        load_balancing:
          exporter:
            insecure: true
          resolver:
            dns:
              hostname: agent-traces-headless.grafana-cloud-monitoring.svc.cluster.local
              port: 34621
              timeout: 5s
              interval: 60s
          receiver_port: 34621
        receivers:
          otlp:
            protocols:
              grpc:
                endpoint: 0.0.0.0:9411
        remote_write:
        - basic_auth:
            username: 111111
            password: pass
          endpoint: tempo-prod-06-prod-gb-south-0.grafana.net:443
          retry_on_failure:
            enabled: false
```
{{< /collapse >}}

{{< collapse title="Example Kubernetes configuration with Kubernetes load balancing" >}}

<!-- TODO: Fill in the Kubernetes yaml once I have a working configuration -->
```yaml
```

{{< /collapse >}}

You need to fill in correct OTLP credentials prior to running the above example.
The example above can be started by using k3d:
<!-- TODO: Link to the k3d page -->
```bash
k3d cluster create grafana-agent-lb-test
kubectl apply -f kubernetes_config.yaml
```

To delete the cluster, run:
```bash
k3d cluster delete grafana-agent-lb-test
```
