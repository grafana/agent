
## Main differences

- We don't handle deployment for you now, you need to use the helm chart to make your own deployments
- If you are using integrations like node-exporter or cadvisor, we recommend you run those outside of the agent now
- You need to make your own config, and pass it into helm chart
- recommended architecture: statefulset / clustering / hpa
- also consider using pete's chart

## Deploy Grafana Agent with Helm

1. You will need to create a `values.yaml` file, which contains options for how to deploy your agent. You may start with the [default values](https://github.com/grafana/agent/blob/main/operations/helm/charts/grafana-agent/values.yaml) and customize as you see fit, or start with this snippet, which should be a good starting point for what the operator does:

    ```yaml
    agent:
      mode: 'flow'
      configMap:
        create: true
      clustering:
        enabled: true
      # if you need to reference secrets for things like
      # authentication, envFrom is a convenient way
      # to mount a secret as environment variables
      envFrom:
        - secretRef:
            name: 'primary-credentials-metrics'
    controller:
      type: 'statefulset'
    replicas: 2
    ```

2. Create a flow config file, `agent.river`. 

3. Install the grafana helm repository:

    ```
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    ```

4. Create a helm relase. You may name the relase anything you like. Here we are installing a release named `grafana-agent` in the `monitoring` namespace.

    ```
    helm upgrade -i -n monitoring -f values.yaml --set-file agent.configMap.content=agent.river
    ```

This command uses the `--set-file` flag to pass the config file as a helm value, so that we can continue to edit it as a regular river file.

## Convert `MetricsIntances` to flow components.

If we have a MetricsInstance like this:

```yaml
apiVersion: monitoring.grafana.com/v1alpha1
kind: MetricsInstance
metadata:
  name: primary
  namespace: monitoring
  labels:
    agent: grafana-agent-metrics
spec:
  remoteWrite:
  - url: your_remote_write_URL
    basicAuth:
      username:
        name: primary-credentials-metrics
        key: username
      password:
        name: primary-credentials-metrics
        key: password
  serviceMonitorNamespaceSelector: {}
  serviceMonitorSelector:
    matchLabels:
      instance: primary

  podMonitorNamespaceSelector: {}
  podMonitorSelector:
    matchLabels:
      instance: primary
  probeNamespaceSelector: {}
  probeSelector:
    matchLabels:
      instance: primary
```

an equivalent river config would be:

```river
prometheus.remote_write "primary" {
    endpoint {
        url = your_remote_write_URL
        basic_auth {
            // these are set as environment variables in the agent pod,
            // from the `primary-credentials-metrics` secret
            // using the `envFrom` value set above
            username = env.username
            password = env.password
        }
    }
}
prometheus.operator.podmonitors "primary" {
    forward_to = [prometheus.remote_write.primary.receiver]
    selector {
        key = "instance"
        operator = "In"
        values = ["primary"]
    }
}
prometheus.operator.servicemonitors "primary" {
    forward_to = [prometheus.remote_write.primary.receiver]
    selector {
        key = "instance"
        operator = "In"
        values = ["primary"]
    }
}
prometheus.operator.probes "primary" {
    forward_to = [prometheus.remote_write.primary.receiver]
    selector {
        key = "instance"
        operator = "In"
        values = ["primary"]
    }
}
```

This config will discover all `PodMonitor`, `ServiceMonitor`, and `Probe` resources in your cluster that match our label selector `instance=primary`. It will then scrape metrics from their targets, and forward them on to your remote write endpoint.

If you are using additional features in your `MetricsInstance` resources, you may need to further customize this config. Please see the documentation for the relevant components fot additional information:

- prometheus.remote_write
- prometheus.operator.podmonitors
- prometheus.operator.servicemonitors
- prometheus.scrape

## Collecting Logs

- LogsInstance -> `loki.write`
- Consider migrating to `loki.source.kubernetes` for simplicity.

## Integrations

- most integrations have equivalent `prometheus.exporter` components.
- If using node_exporter or cadvisor, strongly recommend 