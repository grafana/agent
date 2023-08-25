# Updating OpenTelemetry Collector dependencies

The Agent depends on various OpenTelemetry (Otel) modules such as these:
```
github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter v0.80.0
github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension v0.80.0
github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor v0.80.0
go.opentelemetry.io/collector v0.80.0
go.opentelemetry.io/collector/component v0.80.0
go.opentelemetry.io/otel v1.16.0
go.opentelemetry.io/otel/metric v1.16.0
go.opentelemetry.io/otel/sdk v1.16.0
```

The dependencies mostly come from these repositories:

* [opentelemetry-collector](https://github.com/open-telemetry/opentelemetry-collector)
* [opentelemetry-collector-contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)
* [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go)

Unfortunately, updating Otel dependencies is not straightforward:

* Some of the modules in `opentelemetry-collector` come from a [grafana/opentelemetry-collector](https://github.com/grafana/opentelemetry-collector) fork. 
  * This is mostly so that we can include metrics of Collector components with the metrics shown under the Agent's `/metrics` endpoint.
* All Collector and Collector-Contrib dependencies are kept in sync on the same version.
  * E.g. if we use `v0.80.0` of `go.opentelemetry.io/collector`, we also use `v0.80.0` of `spanmetricsprocessor`.
  * This is in line with how the Collector itself imports dependencies.
  * It helps us avoid bugs.
  * It makes it easier to communicate to customers the version of Collector which we use in the Agent.
  * Unfortunately, updating everything at once makes it tedious to check if any of our docs or code need updating due to changes in Collector components. A lot of these checks are manual - for example, cross checking the Otel config and Otel documentation between versions.
  * There are some exceptions for modules which don't follow the same versioning. For example, `collector/pdata` is usually on a different version, like `v1.0.0-rcv0013`.

## Updating walkthrough

### Update the Grafana fork of Otel Collector

1. Create a new release branch with a `-grafana` suffix under [grafana/opentelemetry-collector](https://github.com/grafana/opentelemetry-collector). For example, if porting branch `v0.81.0`, make a branch under the fork repo called `0.81-grafana`.
2. Check which branch of the fork repo the Agent currently uses.
3. Create a PR to cherry-pick the same changes to the new `-grafana` branch.

### Update the Agent's dependencies

1. Make sure we use the same version of Collector and Collector-Contrib for all relevant modules. For example, if we use version `v0.81.0` of Collector, we should also use version `v0.81.0` for all Contrib modules.
2. Update the `replace` directives in the go.mod file to point to the latest commit of the forked release branch. Use a command like this:
   ```
   go mod edit -replace=go.opentelemetry.io/collector=github.com/grafana/opentelemetry-collector@asdf123jkl
   ```
   Repeat this for any other modules where a replacement is necessary.
3. Note that sometimes Collector depends on packages with "rc" versions such as `v1.0.0-rcv0013`. This is ok, as long as the go.mod of Collector also references the same versions - for example, [pdata](https://github.com/open-telemetry/opentelemetry-collector/blob/v0.81.0/go.mod#L25) and [featuregate](https://github.com/open-telemetry/opentelemetry-collector/blob/v0.81.0/go.mod#L24).

### Update otelcol Flow components

1. Note which Otel components are in use by the Agent.
   * Usually for every "otelcol" Flow component there is a corresponding Collector component.
   * In some cases we don't use the corresponding Collector component:
     * For example, [otelcol.receiver.prometheus](https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.receiver.prometheus/) and [otelcol.exporter.prometheus](https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.exporter.prometheus/).
     * Those components usually have a note like this:
       > NOTE: otelcol.exporter.prometheus is a custom component unrelated to the prometheus exporter from OpenTelemetry Collector.
    * For example, the Otel component used by [otelcol.auth.sigv4](https://grafana.com/docs/agent/latest/flow/reference/components/otelcol.auth.sigv4/) is [sigv4auth](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/sigv4authextension).
2. Make a list of the components which have changed since the previously used version.
   1. Go through the changelogs of both [Collector](https://github.com/open-telemetry/opentelemetry-collector/releases) and [Collector-Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases).
   2. If a component which is in use by the Agent has changed, note it down.
3. For each Otel component which has changed, compare how they changed.
   1. Compare the old and new version of Otel's documentation.
   2. Compare the config.go file to see if new parameters were added.
4. Update the Agent's code and documentation where needed.
   * Pay particular attention to whether the component's stability has changed. 
   Stability is labeled at the top of the Collector documentation. 
   The stability label in Collector's docs should match the one in the Agent's docs.

## Testing

### Testing a tracing pipeline locally

You can use the resources in the [Tempo repository](https://github.com/grafana/tempo/tree/main/example/docker-compose/agent) to create a local source of traces using k6. You can also start your own Tempo and Grafana instances.

1. Comment out the "agent" and "prometheus" sections in the [docker-compose](https://github.com/grafana/tempo/blob/main/example/docker-compose/agent/docker-compose.yaml). We don't need this - instead, we will start our own locally built Agent.
2. Change the "k6-tracing" endpoint to send traces on the localhost, outside of the Docker container.
   * For example, use `ENDPOINT=host.docker.internal:4320`.
   * Then our local Agent should be configured to accept traces on `0.0.0.0:4320`.
3. Optionally, e.g. if you prefer Grafana Cloud, comment out the "tempo" and "grafana" sections of the docker-compose file.
4. Add a second k6 instance if needed - for example, when testing a Static Agent which has 2 Traces instances.

### Static mode

The [tracing subsystem](https://grafana.com/docs/agent/latest/static/configuration/traces-config/) is the only part of Static mode which uses Otel. Try to test as many features of it using a config file like this one:

<details>
  <summary>Example Static config</summary>

```
server:
  log_level: debug

logs:
  positions_directory: "/Users/ExampleUser/Desktop/otel_test/test_log_pos_dir"
  configs:
    - name: "grafanacloud-oteltest-logs"
      clients:
        - url: "https://logs-prod-008.grafana.net/loki/api/v1/push"
          basic_auth:
            username: "USERNAME"
            password: "PASSWORD"

traces:
  configs:
  - name: firstConfig
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: "0.0.0.0:4320"
    remote_write:
      - endpoint: tempo-prod-06-prod-gb-south-0.grafana.net:443
        basic_auth:
          username: "USERNAME"
          password: "PASSWORD"
    batch:
      timeout: 5s
      send_batch_size: 100
    automatic_logging:
      backend: "logs_instance"
      logs_instance_name: "grafanacloud-oteltest-logs"
      roots: true
    spanmetrics:
      handler_endpoint: "localhost:8899"
      namespace: "otel_test_"
    tail_sampling:
      policies:
        [
          {
            name: test-policy-4,
            type: probabilistic,
            probabilistic: {sampling_percentage: 100}
          },
        ]
    service_graphs:
      enabled: true
  - name: secondConfig
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: "0.0.0.0:4321"
    remote_write:
      - endpoint: tempo-prod-06-prod-gb-south-0.grafana.net:443
        basic_auth:
          username: "USERNAME"
          password: "PASSWORD"
    batch:
      timeout: 5s
      send_batch_size: 100
    tail_sampling:
      policies:
        [
          {
            name: test-policy-4,
            type: probabilistic,
            probabilistic: {sampling_percentage: 100}
          },
        ]
    service_graphs:
      enabled: true

```

</details>

Run this file for two types of Agents - an upgraded one, and another one built using the codebase of the `main` branch. Check the following:

* Open `localhost:12345/metrics` in your browser for both Agents.
  * Are new metrics added? Mention them in the changelog.
  * Are metrics missing? Did any metrics change names? If it's intended, mention them in the changelog and the upgrade guide.
* Try opening `localhost:8888/metrics` in your browser for the new Agent. 8888 is the Collector's default port for exposing metrics. Make sure this page doesn't display anything - the Agent should use port `12345` instead.
* Check the logs for errors or anything else that's suspicious.
* Check Tempo to make sure the traces were received.
* Check Loki to make sure the logs generated from traces got received.
* Check `localhost:8899/metrics` to make sure the span metrics are being generated.

### Flow mode

The "otelcol" [components](https://grafana.com/docs/agent/latest/flow/reference/components/) are the only part of Flow mode which uses Otel. Try to test as many of them as possible using a config file like this one:

<details>
  <summary>Example Flow config</summary>

```
otelcol.receiver.otlp "default" {
    grpc {
        endpoint = "0.0.0.0:4320"
    }

    output {
        traces  = [otelcol.processor.batch.default.input]
    }
}

otelcol.processor.batch "default" {
    timeout = "5s"
    send_batch_size = 100

    output {
        traces  = [otelcol.processor.tail_sampling.default.input]
    }
}

otelcol.processor.tail_sampling "default" {
  decision_wait               = "5s"
  num_traces                  = 50000
  expected_new_traces_per_sec = 0

  policy {
    name = "test-policy-1"
    type = "probabilistic"

    probabilistic {
      sampling_percentage = 10
    }
  }

  policy {
    name = "test-policy-2"
    type = "status_code"

    status_code {
      status_codes = ["ERROR"]
    }
  }

  output {
    traces = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
    client {
        endpoint = "localhost:4317"
        tls {
            insecure = true
        }
    }
}
```

</details>

Run this file for two types of Agents - an upgraded one, and another one built using the codebase of the `main` branch. Check the following:

* Open `localhost:12345/metrics` in your browser for both Agents.
  * Are new metrics added? Mention them in the changelog.
  * Are metrics missing? Did any metrics change names? If it's intended, mention them in the changelog and the upgrade guide.
* Check the logs for errors or anything else that's suspicious.
* Check Tempo to make sure the traces were received.
