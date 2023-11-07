---
aliases:
- /docs/grafana-cloud/agent/operator/add-custom-scrape-jobs/
- /docs/grafana-cloud/monitor-infrastructure/agent/operator/add-custom-scrape-jobs/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/operator/add-custom-scrape-jobs/
- /docs/grafana-cloud/send-data/agent/operator/add-custom-scrape-jobs/
canonical: https://grafana.com/docs/agent/latest/operator/add-custom-scrape-jobs/
description: Learn how to add custom scrape jobs
title: Add custom scrape jobs
weight: 400
---

# Add custom scrape jobs

Sometimes you want to add a scrape job for something that isn't supported by the
standard set of Prometheus Operator CRDs. A common example of this is node-level
metrics.

To do this, you'll need to write custom scrape configs and store it in a
Kubernetes Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: extra-jobs
  namespace: operator
stringData:
  jobs.yaml: |
    <SCRAPE CONFIGS>
```

Replace `<SCRAPE CONFIGS>` above with the array of Prometheus scrape jobs to
include.

For example, to collect metrics from Kubelet and cAdvisor, use the following:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: extra-jobs
  namespace: operator
stringData:
  jobs.yaml: |
    - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      job_name: integrations/kubernetes/kubelet
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - replacement: kubernetes.default.svc:443
        target_label: __address__
      - regex: (.+)
        source_labels: [__meta_kubernetes_node_name]
        replacement: /api/v1/nodes/$1/proxy/metrics
        target_label: __metrics_path__
      - action: hashmod
        modulus: $(SHARDS)
        source_labels:
        - __address__
        target_label: __tmp_hash
      - action: keep
        regex: $(SHARD)
        source_labels:
        - __tmp_hash
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      job_name: integrations/kubernetes/cadvisor
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - replacement: kubernetes.default.svc:443
        target_label: __address__
      - regex: (.+)
        replacement: /api/v1/nodes/$1/proxy/metrics/cadvisor
        source_labels:
        - __meta_kubernetes_node_name
        target_label: __metrics_path__
      - action: hashmod
        modulus: $(SHARDS)
        source_labels:
        - __address__
        target_label: __tmp_hash
      - action: keep
        regex: $(SHARD)
        source_labels:
        - __tmp_hash
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
```

Note that you **should** always add these two relabel_configs for each custom job:

```yaml
- action: hashmod
  modulus: $(SHARDS)
  source_labels:
  - __address__
  target_label: __tmp_hash
- action: keep
  regex: $(SHARD)
  source_labels:
  - __tmp_hash
```

These rules ensure if your GrafanaAgent has multiple metrics shards, only one
pod per replica will collect metrics for each job.

Once your Secret is defined, you'll then need to add a `additionalScrapeConfigs`
field to your MetricsInstance:

```yaml
apiVersion: monitoring.grafana.com/v1alpha1
kind: MetricsInstance
metadata:
  labels:
    name: grafana-agent
  name: primary
  namespace: operator
spec:
  additionalScrapeConfigs:
    name: extra-jobs
    key: jobs.yaml
  # ... Other settings ...
```

The Secret **MUST** be in the same namespace as the MetricsInstance.
