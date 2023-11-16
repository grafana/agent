---
description: Learn how to Flow compares to the OpenTelemetry Collector
menuTitle: OpenTelemetry Collector
title: OpenTelemetry Collector
weight: 100
---

# Comparing Flow with the OpenTelemetry Collector

## Which one of the two should you use? 

Flow may be a better choice if:
* You prefer River syntax over YAML.
* You would like to collect "profile" telemetry - Flow can do this via Pyroscope components.
* You would like to process Prometheus or Loki telemetry without converting it to OTLP.

The Collector may be a better choice if:
* You are already using the Collector, and do not need any of the features of Flow.
* You prefer YAML syntax over River.

Below you can find more details on similarities and differences between the Collector and Flow.

## Similarities

### Many Collector components are available in Flow

Flow contains many of the components available inside the Collector:
* For example, the OTLP exporter is available as `otelcol.exporter.otlp`.
* The Grafana Agent development team is actively working on adding more of the Collector's components to Flow.
* If a Flow user needs a Collector feature to be made available in the Agent, the Agent development team could work on implementing it.

### Similar performance when processing OpenTelemetry signals natively

Most of Flow's `otelcol` components are just thin wrappers over a Collector component.
Hence, the CPU and memory performance of Flow is similar.

## Differences

### Configuration language

Collector is configured using yaml, whereas Flow is configured using River.

#### Example - coalesce

One of River's main advantages is its standard library. It contains handy functions such as coalesce, 
which could be used to retrieve the first argument which is not null or empty:

```river
```

### Modularity

The Agent configuration is more flexible, modular, and allows for more opportunities to chain components together in a pipeline. 

#### Example - retrieving data from a file

Let's say you would like to use OAuth2 authentication in the Collector. If you need to retrieve `client_id` 
or `client_secret` from a file, then you would have to use the `client_id_file` or `client_secret_file` config parameters.

```yaml
```

In the Agent, you'd use `otelcol.auth.oauth2` with the normal `client_id` and `client_secret` parameters, 
and you would setup another component which retrieves those from a `local.file` component. 

```river
```

Moreover, the string could also come from a `remote.s3`, `remote.vault`, or `remote.http`. 
This gives Flow users lots of flexibility because strings coming out of those components 
could be used for any parameter, in any component which requires a string - not just for 
a `client_id` for OAuth2.

```river
```

### Flow can process Prometheus signals natively

Collector needs to convert Prometheus signals to the OTLP format in order to process them.
Flow, on the other hand, can process those signals natively using components such as `prometheus.relabel`, `prometheus.relabel`, and `prometheus.remote_write`.
This could lead to better performance and ease of use.

### Flow documentation is consistent and structured

Flow components tend to be documented in a more consistent way than Collector components. 

### Some Collector features are not available in the Agent and vice-versa

Agent doesn't have all the components which the Collector has. However, the Grafana  Agent development team is working on 
adding new components all the time and we would be happy to add new components which Flow users need.

### Flow supports "profile" telemetry signals

OpenTelemetry currently does not support "profile" signals. Flow supports them through components such as `pyroscope.scrape` and `pyroscope.ebpf`.

### Flow is usually a few versions of OpenTelemetry behind

Flow imports OpenTelemetry libraries as a dependency. Generally, a given version of Flow might use OpenTelemetry libraries 
which are older by 1, 2, or even 4 months. This is usually not a problem and the Agent development team could schedule an 
upgrade of the Agent's OpenTelemetry dependencies outside of the usual upgrade cycle if it is important to a user.
