---
canonical: https://grafana.com/docs/agent/latest/flow/getting-started/collect-prometheus-metrics/
title: Collect Prometheus metrics
weight: 200
---

# Collect Prometheus metrics

Grafana Agent Flow can be configured to collect [Prometheus][] metrics and
forward them to any Prometheus-compatible database.

This topic describes how to:

* Configure metrics delivery
* Collect metrics from Kubernetes Pods

[Prometheus]: https://prometheus.io

## Components used in this topic

* [discovery.kubernetes][]
* [prometheus.remote_write][]
* [prometheus.scrape][]

[discovery.kubernetes]: {{< relref "../reference/components/discovery.kubernetes.md" >}}
[prometheus.remote_write]: {{< relref "../reference/components/prometheus.remote_write.md" >}}
[prometheus.scrape]: {{< relref "../reference/components/prometheus.scrape.md" >}}

## Before you begin

* Ensure that you have basic familiarity with instrumenting applications with
  Prometheus.
* Have a set of Prometheus exports or applications exposing Prometheus metrics
  that you want to collect metrics from.
* Identify where you will write collected metrics. Metrics may be written to
  Prometheus or Prometheus-compatible endpoints such as Grafana Mimir, Grafana
  Cloud, or Grafana Enterprise Metrics.
* Be familiar with the concept of [Components][] in Grafana Agent Flow.

[Components]: {{< relref "../concepts/components.md" >}}

## Configure metrics delivery

Before components can collect Prometheus metrics, you must have a component
responsible for writing those metrics somewhere.

The [prometheus.remote_write][] component is responsible for delivering
Prometheus metrics to one or Prometheus-compatible endpoints. Once a
`prometheus.remote_write` component is defined, other Grafana Agent Flow
components can be used to forward metrics to it.

To configure a `prometheus.remote_write` component for metrics delivery,
complete the following steps:

1. Add the following `prometheus.remote_write` component to your configuration file:

   ```river
   prometheus.remote_write "LABEL" {
     endpoint {
       url = "PROMETHEUS_URL"
     }
   }
   ```

    1. Replace `LABEL` with a label to use for the component, such as `default`.
       The label chosen must be unique across all `prometheus.remote_write`
       components in the same configuration file.

    2. Replace `PROMETHEUS_URL` with the full URL of the Prometheus-compatible
       endpoint where metrics will be sent, such as
       `https://prometheus-us-central1.grafana.net/api/prom/push`.

2. If your endpoint requires basic authentication, paste the following inside
   of the `endpoint` block:

   ```river
   basic_auth {
     username = "USERNAME"
     password = "PASSWORD"
   }
   ```

    1. Replace `USERNAME` with the basic authentication username to use.

    2. Replace `PASSWORD` with the basic authentication password or API key to
       use.

3. If you have more than one endpoint to write metrics to, repeat the
   `endpoint` block for additional endpoints.

The following example demonstrates configuring `prometheus.remote_write` with
multiple endpoints and mixed usage of basic authentication, and a
`prometheus.scrape` component which forwards metrics to it:

```river
prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9090/api/prom/push"
  }

  endpoint {
    url = "https://prometheus-us-central1.grafana.net/api/prom/push"

    // Get basic authentication based on environment variables.
    basic_auth {
      username = env("REMOTE_WRITE_USERNAME")
      password = env("REMOTE_WRITE_PASSWORD")
    }
  }
}

prometheus.scrape "example" {
  // Collect metrics from Grafana Agent's default listen address.
  targets = [{
    __address__ = "127.0.0.1:12345",
  }]

  forward_to = [prometheus.remote_write.default.receiver]
}
```

For more information on configuring metrics delivery, refer to
[prometheus.remote_write][].

## Collect metrics from Kubernetes Pods

Grafana Agent Flow can be configured to collect metrics from Kubernetes Pods
by:

1. Discovering Kubernetes Pods to collect metrics from.
2. Collecting metrics from those discovered Pods.

To collect metrics from Kubernetes Pods, complete the following steps:

1. Follow [Configure metrics delivery](#configure-metrics-delivery) to ensure
   collected metrics can be written somewhere.

2. Discover Kubernetes Pods:

    1. Add the following `discovery.kubernetes` component to your configuration
       file to discover every Pod in the cluster across all Namespaces:

       ```river
       discovery.kubernetes "DISCOVERY_LABEL" {
         role = "pod"
       }
       ```

       This will generate one Prometheus target for every exposed port on every
       discovered Pod.

        1. Replace `DISCOVERY_LABEL` with a label to use for the component, such as
           `pods`. The label chosen must be unique across all
           `discovery.kubernetes` components in the same configuration file.

    2. To limit the Namespaces that Pods are discovered in, add the following
       block inside of the `discovery.kubernetes` component:

       ```river
       namespaces {
         own_namespace = true
         names         = [NAMESPACE_NAMES]
       }
       ```

        1. If you do not want to search for Pods in the Namespace Grafana
           Agent is running in, set `own_namespace` to `false`.

        2. Replace `NAMESPACE_NAMES` with a comma-delimited list of strings
           representing Namespaces to search. Each string must be wrapped in
           double quotes. For example, `"default","kube-system"`.

    3. To use a field selector to limit the number of discovered Pods, add the
       following block inside of the `discovery.kubernetes` component:

       ```river
       selectors {
         role  = "pod"
         field = "FIELD_SELECTOR"
       }
       ```

        1. Replace `FIELD_SELECTOR` with the Kubernetes field selector to use,
           such as `metadata.name=my-service`. For more information on field
           selectors, refer to the Kubernetes documentation on [Field
           Selectors][].

        2. Create additional `selectors` blocks for each field selector you
           want to apply.

    4. To use a label selector to limit the number of discovered Pods, add the
       following block inside of the `discovery.kubernetes` component:

       ```river
       selectors {
         role  = "pod"
         label = "LABEL_SELECTOR"
       }
       ```

        1. Replace `LABEL_SELECTOR` with the Kubernetes label selector to use,
           such as `environment in (production, qa)`. For more information on
           label selectors, refer to the Kubernetes documentation on [Labels
           and Selectors][].

        2. Create additional `selectors` blocks for each label selector you
           want to apply.

3. Collect metrics from discovered Pods:

    1. Add the following `prometheus.scrape` component to your configuration
       file:

       ```river
       prometheus.scrape "SCRAPE_LABEL" {
         targets    = discovery.kubernetes.DISCOVERY_LABEL.targets
         forward_to = [prometheus.remote_write.REMOTE_WRITE_LABEL.receiver]
       }
       ```

        1. Replace `SCRAPE_LABEL` with a label to use for the component, such
           as `pods`. The label chosen must be unique across all
           `prometeus.scrape` components in the same configuration file.

        2. Replace `DISCOVERY_LABEL` with the label chosen for the
           `discovery.kubernetes` component in step 2.1.1.

        3. Replace `REMOTE_WRITE_LABEL` with the label chosen for your existing
           `prometheus.remote_write` component.


[Field Selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
[Labels and Selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#set-based-requirement

The following example demonstrates configuring Grafana Agent to collect metrics
from running production Kubernetes Pods in the `default` Namespace:

```river
discovery.kubernetes "pods" {
  role = "pod"

  namespaces {
    own_namespace = false

    names = ["default"]
  }

  selectors {
    role  = "pod"
    label = "environment in (production)"
  }
}

prometheus.scrape "pods" {
  targets    = discovery.kubernetes.pods.targets
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9090/api/prom/push"
  }
}
```

For more information on configuring Kubernetes service delivery and collecting
metrics, refer to [discovery.kubernetes][] and [prometheus.scrape][].

## Collect metrics from Kubernetes Services

Grafana Agent Flow can be configured to collect metrics from Kubernetes Services
by:

1. Discovering Kubernetes Services to collect metrics from.
2. Collecting metrics from those discovered Services.

To collect metrics from Kubernetes Services, complete the following steps:

1. Follow [Configure metrics delivery](#configure-metrics-delivery) to ensure
   collected metrics can be written somewhere.

2. Discover Kubernetes Services:

    1. Add the following `discovery.kubernetes` component to your configuration
       file to discover every Services in the cluster across all Namespaces:

       ```river
       discovery.kubernetes "DISCOVERY_LABEL" {
         role = "service"
       }
       ```

       This will generate one Prometheus target for every exposed port on every
       discovered Service.

        1. Replace `DISCOVERY_LABEL` with a label to use for the component, such as
           `services`. The label chosen must be unique across all
           `discovery.kubernetes` components in the same configuration file.

    2. To limit the Namespaces that Services are discovered in, add the following
       block inside of the `discovery.kubernetes` component:

       ```river
       namespaces {
         own_namespace = true
         names         = [NAMESPACE_NAMES]
       }
       ```

        1. If you do not want to search for Services in the Namespace Grafana
           Agent is running in, set `own_namespace` to `false`.

        2. Replace `NAMESPACE_NAMES` with a comma-delimited list of strings
           representing Namespaces to search. Each string must be wrapped in
           double quotes. For example, `"default","kube-system"`.

    3. To use a field selector to limit the number of discovered Services, add the
       following block inside of the `discovery.kubernetes` component:

       ```river
       selectors {
         role  = "service"
         field = "FIELD_SELECTOR"
       }
       ```

        1. Replace `FIELD_SELECTOR` with the Kubernetes field selector to use,
           such as `metadata.name=my-service`. For more information on field
           selectors, refer to the Kubernetes documentation on [Field
           Selectors][].

        2. Create additional `selectors` blocks for each field selector you
           want to apply.

    4. To use a label selector to limit the number of discovered Services, add the
       following block inside of the `discovery.kubernetes` component:

       ```river
       selectors {
         role  = "service"
         label = "LABEL_SELECTOR"
       }
       ```

        1. Replace `LABEL_SELECTOR` with the Kubernetes label selector to use,
           such as `environment in (production, qa)`. For more information on
           label selectors, refer to the Kubernetes documentation on [Labels
           and Selectors][].

        2. Create additional `selectors` blocks for each label selector you
           want to apply.

3. Collect metrics from discovered Services:

    1. Add the following `prometheus.scrape` component to your configuration
       file:

       ```river
       prometheus.scrape "SCRAPE_LABEL" {
         targets    = discovery.kubernetes.DISCOVERY_LABEL.targets
         forward_to = [prometheus.remote_write.REMOTE_WRITE_LABEL.receiver]
       }
       ```

        1. Replace `SCRAPE_LABEL` with a label to use for the component, such
           as `services`. The label chosen must be unique across all
           `prometeus.scrape` components in the same configuration file.

        2. Replace `DISCOVERY_LABEL` with the label chosen for the
           `discovery.kubernetes` component in step 2.1.1.

        3. Replace `REMOTE_WRITE_LABEL` with the label chosen for your existing
           `prometheus.remote_write` component.


[Field Selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
[Labels and Selectors]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#set-based-requirement

The following example demonstrates configuring Grafana Agent to collect metrics
from running production Kubernetes Services in the `default` Namespace:

```river
discovery.kubernetes "services" {
  role = "service"

  namespaces {
    own_namespace = false

    names = ["default"]
  }

  selectors {
    role  = "service"
    label = "environment in (production)"
  }
}

prometheus.scrape "services" {
  targets    = discovery.kubernetes.services.targets
  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9090/api/prom/push"
  }
}
```

For more information on configuring Kubernetes service delivery and collecting
metrics, refer to [discovery.kubernetes][] and [prometheus.scrape][].

## Collect metrics from custom targets

Grafana Agent Flow can be configured to collect metrics from a custom set of
targets without the need for service discovery.

To collect metrics from a custom set of targets, complete the following steps:

1. Follow [Configure metrics delivery](#configure-metrics-delivery) to ensure
   collected metrics can be written somewhere.

2. Add the following `prometheus.scrape` component to your configuration file:

   ```river
   prometheus.scrape "SCRAPE_LABEL" {
     targets    = [TARGET_LIST]
     forward_to = [prometheus.remote_write.REMOTE_WRITE_LABEL.receiver]
   }
   ```

     1. Replace `SCRAPE_LABEL` with a label to use for the component, such as
        `custom_targets`. The label chosen must be unique across all
        `prometheus.scrape` components in the same configuration file.

     2. Replace `TARGET_LIST` with a comma-delimited list of [Objects][]
        denoting the Prometheus target. Each object must conform to the
        following rules:

        * There must be an `__address__` key denoting the `HOST:PORT` of the
          target to collect metrics from.

        * To explicitly specify which protocol to use, set the `__scheme__` key
          to `"http"` or `"https"`. If the `__scheme__` key is not provided,
          the protocol to use is inherited by the settings of the
          `prometheus.scrape` component (default `"http"`).

        * To explicitly specify which HTTP path to collect metrics from, set
          the `__metrics_path__` key to the HTTP path to use. If the
          `__metrics_path__` key is not provided, the protocol to use is
          inherited by the settings of the `prometheus.scrape` component
          (default `"/metrics"`).

        * Add additional keys as desired to inject extra labels to collected
          metrics. Any label starting with two underscores (`__`) will be
          dropped prior to scraping.

     3. Replace `REMOTE_WRITE_LABEL` with the label chosen for your existing
        `prometheus.remote_write` component.

The following example demonstrates configuring `prometheus.scrape` to collect
metrics from a custom set of endpoints:

```river
prometheus.scrape "custom_targets" {
  targets = [
    {
      __address__ = "prometheus:9090",
    },
    {
      __address__ = "mimir:8080",
      __scheme__  = "https",
    },
    {
      __address__      = "custom-application:80",
      __metrics_path__ = "/custom-metrics–path",
    },
    {
      __address__ = "grafana-agent:12345",
      application = "grafana-agent",
      environment = "production",
    },
  ]

  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = "http://localhost:9090/api/prom/push"
  }
}
```

[Objects]: {{< relref "../config-language/expressions/types_and_values.md#objects" >}}
