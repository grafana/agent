
## Migrating from Grafana Agent Operator

With the release of flow, Grafana Agent Operator is no longer the recommended way to deploy Grafana Agent. Some of the Operator functionality has been moved into Grafana Agent
itself, and the remaining functionality has been replaced by our Helm Chart.

- The Monitor types (`PodMonitor`, `ServiceMonitor`, `Probe`, and `LogsInstance`) are all supported natively by Grafana Agent in Flow mode. You no longer are 
required to use the Operator to consume those CRDs for dynamic monitoring in your cluster.
- The parts of the Operator that deploy the agent itself (`GrafanaAgent`, `MetricsInstance`, and `LogsInstance` CRDs) are depricated. We now recommend
operator users use the [Grafana Agent Helm Chart](https://grafana.com/docs/agent/latest/flow/setup/install/kubernetes/) to deploy the Agent directly to your clusters.

This guide will provide some steps to get started with Grafana Agent for users coming from Grafana Agent Operator.

## Deploy Grafana Agent with Helm

1. You will need to create a `values.yaml` file, which contains options for how to deploy your agent. You may start with the [default values](https://github.com/grafana/agent/blob/main/operations/helm/charts/grafana-agent/values.yaml) and customize as you see fit, or start with this snippet, which should be a good starting point for what the operator does:

    ```yaml
    agent:
      mode: 'flow'
      configMap:
        create: true
      clustering:
        enabled: true
    controller:
      type: 'statefulset'
    replicas: 2
    ```

  This config will use Grafana Agent's built in [clustering]() to allow distributing scrapes across all Agent Pods.

2. Create a flow config file, `agent.river`.

You can add any config you need directly to this file.

3. Install the grafana helm repository:

    ```
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    ```

4. Create a helm relase. You may name the relase anything you like. Here we are installing a release named `grafana-agent` in the `monitoring` namespace.

    ```
    helm upgrade grafana-agent grafana/grafana-agent -i -n monitoring -f values.yaml --set-file agent.configMap.content=agent.river
    ```

This command uses the `--set-file` flag to pass the config file as a helm value, so that we can continue to edit it as a regular river file.

## Convert `MetricsIntances` to flow components.

A `MetricsInstance` resource primarily defines:

- The remote endpoint(s) Grafana Agent to send metrics to.
- Which `PodMonitor`, `ServiceMonitor`, and `Probe` resources this Agent should discover.

These functions can be done in Grafana Agent Flow with the `prometheus.remote_write`, `prometheus.operator.podmonitors`, `prometheus.operator.servicemonitors`, and `prometheus.operator.probes` components respectively.

This is a river sample that is equivalent to the `MetricsInstance` from our [operator guide](https://grafana.com/docs/agent/latest/operator/deploy-agent-operator-resources/#deploy-a-metricsinstance-resource):

```river

// read the credentials secret for remote_write authorization
remote.kubernetes.secret "credentials" {
  namespace = "monitoring"
  name = "primary-credentials-metrics"
}

prometheus.remote_write "primary" {
    endpoint {
        url = your_remote_write_URL
        basic_auth {
            username = nonsensitive(remote.kubernetes.secret.credentials.data["username"])
            password = remote.kubernetes.secret.credentials.data["password"]
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

- [remote.kubernetes.secret](https://grafana.com/docs/agent/latest/flow/reference/components/remote.kubernetes.secret)
- [prometheus.remote_write](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.remote_write)
- [prometheus.operator.podmonitors](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.operator.podmonitors)
- [prometheus.operator.servicemonitors](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.operator.servicemonitors)
- [prometheus.operator.probes](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.operator.probes)
- [prometheus.scrape](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.scrape)

## Collecting Logs

To Be Continued

## Integrations

To Be Continued