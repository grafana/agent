---
title: Kuberenetes Deployment Modes
---
## StatefulSet

A `StatefulSet` is the most common way Grafana Agent is deployed in Kubernetes. Advantages include:

- A relatively small number of Agents can handle metrics for an entire cluster.
- Persistent volumes can be attatched to each pod to persist agent data like the write-ahead log or other component state.
- Number of pods can be scaled up or down as needed, either manually or with horizontal pod autoscaling.

### Distributing work with StatefulSets

Grafana Agent's [built-in clustering](https://grafana.com/docs/agent/latest/flow/getting-started/distribute-prometheus-scrape-load/) is the easiest way to distribute scrape load across a `StatefulSet` of Agents. 

## Deployment

A `Deployment` has most of the same properties as a `StatefulSet`, but cannot use Persistent Volume Claim Templates. It may be desired if you do not want storage, but most of the other properties are the same as a `StatefulSet`.

## DaemonSet

A `DaemonSet` will deploy a Grafana Agent pod to each Node in your Kubernetes cluster. This is needed when you need access to something that is present on every Node in your cluster, such as:

- Local access to log files.
- Exporters, such as `node_exporter`, which need to run on every physical node.

There are several downsides to a `DaemonSet` which need to be considered:

- You cannot attach Persistent Volumes to `DaemonSet` pods. This means things like the Write-Ahead Log will use the Node's local disk, which may not be desired in shared environments.

- Each Agent Pod with remote write components will have its' own Write-Ahead Log. This could increase the total memory used across all agents, especially if there is a temporary disruption writing to the backend storage.

### Distributing work in a DaemonSet

// TODO: should this go in the page on "distributing prometheus scrape"?

In order to avoid scraping targets multiple times, it is best to have each Grafana Agent in a DaemonSet scrape targets only on its' local Node. The most performant way to do this is to apply a field selector to your `discovery.kubernetes` components that discover pods or endpoints:

```river
discovery.kubernetes "pods" {
    role = "pod"
    selectors {
        role = "pod"
        field = "spec.nodeName=" + constants.hostname
    }
}
discovery.kubernetes "endpoints" {
    role = "endpoint"
    selectors {
        role = "pod"
        field = "spec.nodeName=" + constants.hostname
    }
}
```

Targets discovered like this can be filtered with `discovery.relabel` components and scraped with `prometheus.scrape` components without clustering enabled. 

// todo: maybe logging here should also be another page.

Pod logs can be scraped with an additional `discovery.relabel` and `loki.source.file` component:

```river
discovery.relabel "pod_logs" {
  targets = discovery.kubernetes.pods.targets
  rule {
    source_labels = ["__meta_kubernetes_pod_uid", "__meta_kubernetes_pod_container_name"]
    separator = "/"
    action = "replace"
    replacement = "/var/log/pods/*$1/*.log"
    target_label = "__path__"
  }
  // additional rules to filter or add labels to pod logs can go here
}
local.file_match "pod_logs" {
  path_targets = discovery.relabel.pod_logs.output
}
loki.source.file "pod_logs" {
  targets    = local.file_match.pod_logs.targets
  forward_to = [] //Add loki.process or loki.remote_write here
}
```

Some targets are not inherently "node-based", and cannot be distributed by simply filtering by node name like the above samples. Examples of such cases may be:

- "Black-box" style monitoring of services, or ingresses. 
- Certain `prometheus.exporter.*` components that collect data about centralized databases or resources.

To distribute these scrape targets between agents, the [built-in clustering](https://grafana.com/docs/agent/latest/flow/getting-started/distribute-prometheus-scrape-load/) can be utilized:

```river
discovery.kubernetes "services" {
    role = "service"
}
discovery.relabel "filtered_services" {
   targets = discovery.kubernetes.services.targets
   // only scrape services labelled with "scrape=true"
   rule {
     action = "keep"
     source_labels = ["__meta_kubernetes_service_labelpresent_scrape","__meta_kubernetes_service_label_scrape"]
     regex = "true;true"
   }
   // other relabel rules to add labels or filter further
}
prometheus.exporter.redis "redis" {
  // TODO: this works with clustering almost incidentally.
  // the target will be a local address, but clustering will make sure only one agent
  // scrapes it still. We may want to wait for a stronger mechanism of disabling components
  // to advise this
  redis_addr = "redis.redis:6379"
}
prometheus.scrape "demo" {
  targets    = concat(prometheus.exporter.redis.redis.targets,discovery.relabel.filtered_services.targets)
  forward_to = [] // add remote_write here
  clustering {
    enabled = true
  }
}
```




