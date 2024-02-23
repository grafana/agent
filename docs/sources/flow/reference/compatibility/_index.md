---
aliases:
- /docs/grafana-cloud/agent/flow/reference/compatible-components/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/reference/compatible-components/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/reference/compatible-components/
- /docs/grafana-cloud/send-data/agent/flow/reference/compatible-components/
canonical: https://grafana.com/docs/agent/latest/flow/reference/compatibility/
description: Learn about which components are compatible with each other in Grafana Agent Flow
title: Compatible components
weight: 400
---

# Compatible components

This section provides an overview of _some_ of the possible connections between compatible components in {{< param "PRODUCT_NAME" >}}.

For each common data type, we provide a list of compatible components that can export or consume it.

{{< admonition type="note" >}}
The type of export may not be the only requirement for chaining components together.
The value of an attribute may matter as well as its type.
Refer to each component's documentation for more details on what values are acceptable.

For example:
* A Prometheus component may always expect an `"__address__"` label inside a list of targets.
* A `string` argument may only accept certain values like "traceID" or "spanID".
{{< /admonition >}}

## Targets

Targets are a `list(map(string))` - a [list]({{< relref "../../concepts/config-language/expressions/types_and_values/#naming-convention" >}}) of [maps]({{< relref "../../concepts/config-language/expressions/types_and_values/#naming-convention" >}}) with [string]({{< relref "../../concepts/config-language/expressions/types_and_values/#strings" >}}) values.
They can contain different key-value pairs, and you can use them with a wide range of
components. Some components require Targets to contain specific key-value pairs
to work correctly. It's recommended to always check component references for
details when working with Targets.

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Targets Exporters
The following components, grouped by namespace, _export_ Targets.

<!-- START GENERATED SECTION: EXPORTERS OF Targets -->

{{< collapse title="discovery" >}}
- [discovery.azure](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.azure)
- [discovery.consul](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.consul)
- [discovery.consulagent](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.consulagent)
- [discovery.digitalocean](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.digitalocean)
- [discovery.dns](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.dns)
- [discovery.docker](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.docker)
- [discovery.dockerswarm](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.dockerswarm)
- [discovery.ec2](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.ec2)
- [discovery.eureka](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.eureka)
- [discovery.file](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.file)
- [discovery.gce](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.gce)
- [discovery.hetzner](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.hetzner)
- [discovery.http](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.http)
- [discovery.ionos](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.ionos)
- [discovery.kubelet](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.kubelet)
- [discovery.kubernetes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.kubernetes)
- [discovery.kuma](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.kuma)
- [discovery.lightsail](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.lightsail)
- [discovery.linode](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.linode)
- [discovery.marathon](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.marathon)
- [discovery.nerve](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.nerve)
- [discovery.nomad](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.nomad)
- [discovery.openstack](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.openstack)
- [discovery.ovhcloud](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.ovhcloud)
- [discovery.process](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.process)
- [discovery.puppetdb](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.puppetdb)
- [discovery.relabel](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.relabel)
- [discovery.scaleway](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.scaleway)
- [discovery.serverset](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.serverset)
- [discovery.triton](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.triton)
- [discovery.uyuni](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.uyuni)
{{< /collapse >}}

{{< collapse title="local" >}}
- [local.file_match](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/local.file_match)
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.exporter.apache](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.apache)
- [prometheus.exporter.azure](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.azure)
- [prometheus.exporter.blackbox](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.blackbox)
- [prometheus.exporter.cadvisor](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.cadvisor)
- [prometheus.exporter.cloudwatch](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.cloudwatch)
- [prometheus.exporter.consul](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.consul)
- [prometheus.exporter.dnsmasq](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.dnsmasq)
- [prometheus.exporter.elasticsearch](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.elasticsearch)
- [prometheus.exporter.gcp](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.gcp)
- [prometheus.exporter.github](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.github)
- [prometheus.exporter.kafka](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.kafka)
- [prometheus.exporter.memcached](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.memcached)
- [prometheus.exporter.mongodb](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.mongodb)
- [prometheus.exporter.mssql](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.mssql)
- [prometheus.exporter.mysql](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.mysql)
- [prometheus.exporter.oracledb](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.oracledb)
- [prometheus.exporter.postgres](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.postgres)
- [prometheus.exporter.process](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.process)
- [prometheus.exporter.redis](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.redis)
- [prometheus.exporter.self](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.self)
- [prometheus.exporter.snmp](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.snmp)
- [prometheus.exporter.snowflake](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.snowflake)
- [prometheus.exporter.squid](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.squid)
- [prometheus.exporter.statsd](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.statsd)
- [prometheus.exporter.unix](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.unix)
- [prometheus.exporter.vsphere](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.vsphere)
- [prometheus.exporter.windows](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.exporter.windows)
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Targets -->


<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Targets Consumers
The following components, grouped by namespace, _consume_ Targets.

<!-- START GENERATED SECTION: CONSUMERS OF Targets -->

{{< collapse title="discovery" >}}
- [discovery.process](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.process)
- [discovery.relabel](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/discovery.relabel)
{{< /collapse >}}

{{< collapse title="local" >}}
- [local.file_match](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/local.file_match)
{{< /collapse >}}

{{< collapse title="loki" >}}
- [loki.source.docker](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.docker)
- [loki.source.file](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.file)
- [loki.source.kubernetes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.kubernetes)
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.processor.discovery](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.discovery)
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.scrape](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.scrape)
{{< /collapse >}}

{{< collapse title="pyroscope" >}}
- [pyroscope.ebpf](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.ebpf)
- [pyroscope.java](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.java)
- [pyroscope.scrape](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.scrape)
{{< /collapse >}}

<!-- END GENERATED SECTION: CONSUMERS OF Targets -->


## Prometheus `MetricsReceiver`

The Prometheus metrics are sent between components using `MetricsReceiver`s.
`MetricsReceiver`s are [capsules]({{< relref "../../concepts/config-language/expressions/types_and_values/#capsules" >}})
that are exported by components that can receive Prometheus metrics. Components that
can consume Prometheus metrics can be passed the `MetricsReceiver` as an argument. Use the
following components to build your Prometheus metrics pipeline:

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Prometheus `MetricsReceiver` Exporters
The following components, grouped by namespace, _export_ Prometheus `MetricsReceiver`.

<!-- START GENERATED SECTION: EXPORTERS OF Prometheus `MetricsReceiver` -->

{{< collapse title="otelcol" >}}
- [otelcol.receiver.prometheus](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.prometheus)
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.relabel](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.relabel)
- [prometheus.remote_write](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.remote_write)
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Prometheus `MetricsReceiver` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Prometheus `MetricsReceiver` Consumers
The following components, grouped by namespace, _consume_ Prometheus `MetricsReceiver`.

<!-- START GENERATED SECTION: CONSUMERS OF Prometheus `MetricsReceiver` -->

{{< collapse title="otelcol" >}}
- [otelcol.exporter.prometheus](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.prometheus)
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.operator.podmonitors](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.operator.podmonitors)
- [prometheus.operator.probes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.operator.probes)
- [prometheus.operator.servicemonitors](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.operator.servicemonitors)
- [prometheus.receive_http](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.receive_http)
- [prometheus.relabel](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.relabel)
- [prometheus.scrape](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/prometheus.scrape)
{{< /collapse >}}

<!-- END GENERATED SECTION: CONSUMERS OF Prometheus `MetricsReceiver` -->

## Loki `LogsReceiver`

`LogsReceiver` is a [capsule]({{< relref "../../concepts/config-language/expressions/types_and_values/#capsules" >}})
that is exported by components that can receive Loki logs. Components that
consume `LogsReceiver` as an argument typically send logs to it. Use the
following components to build your Loki logs pipeline:

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Loki `LogsReceiver` Exporters
The following components, grouped by namespace, _export_ Loki `LogsReceiver`.

<!-- START GENERATED SECTION: EXPORTERS OF Loki `LogsReceiver` -->

{{< collapse title="loki" >}}
- [loki.echo](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.echo)
- [loki.process](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.process)
- [loki.relabel](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.relabel)
- [loki.write](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.write)
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.receiver.loki](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.loki)
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Loki `LogsReceiver` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Loki `LogsReceiver` Consumers
The following components, grouped by namespace, _consume_ Loki `LogsReceiver`.

<!-- START GENERATED SECTION: CONSUMERS OF Loki `LogsReceiver` -->

{{< collapse title="faro" >}}
- [faro.receiver](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/faro.receiver)
{{< /collapse >}}

{{< collapse title="loki" >}}
- [loki.process](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.process)
- [loki.relabel](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.relabel)
- [loki.source.api](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.api)
- [loki.source.awsfirehose](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.awsfirehose)
- [loki.source.azure_event_hubs](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.azure_event_hubs)
- [loki.source.cloudflare](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.cloudflare)
- [loki.source.docker](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.docker)
- [loki.source.file](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.file)
- [loki.source.gcplog](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.gcplog)
- [loki.source.gelf](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.gelf)
- [loki.source.heroku](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.heroku)
- [loki.source.journal](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.journal)
- [loki.source.kafka](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.kafka)
- [loki.source.kubernetes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.kubernetes)
- [loki.source.kubernetes_events](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.kubernetes_events)
- [loki.source.podlogs](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.podlogs)
- [loki.source.syslog](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.syslog)
- [loki.source.windowsevent](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/loki.source.windowsevent)
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.exporter.loki](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.loki)
{{< /collapse >}}

<!-- END GENERATED SECTION: CONSUMERS OF Loki `LogsReceiver` -->

## OpenTelemetry `otelcol.Consumer`

The OpenTelemetry data is sent between components using `otelcol.Consumer`s.
`otelcol.Consumer`s are [capsules]({{< relref "../../concepts/config-language/expressions/types_and_values/#capsules" >}})
that are exported by components that can receive OpenTelemetry data. Components that
can consume OpenTelemetry data can be passed the `otelcol.Consumer` as an argument. Note that some components
that use `otelcol.Consumer` only support a subset of telemetry signals, for example, only traces. Check the component
reference pages for more details on what is supported. Use the following components to build your OpenTelemetry pipeline:

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### OpenTelemetry `otelcol.Consumer` Exporters
The following components, grouped by namespace, _export_ OpenTelemetry `otelcol.Consumer`.

<!-- START GENERATED SECTION: EXPORTERS OF OpenTelemetry `otelcol.Consumer` -->

{{< collapse title="otelcol" >}}
- [otelcol.connector.host_info](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.host_info)
- [otelcol.connector.servicegraph](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.servicegraph)
- [otelcol.connector.spanlogs](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.spanlogs)
- [otelcol.connector.spanmetrics](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.spanmetrics)
- [otelcol.exporter.loadbalancing](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.loadbalancing)
- [otelcol.exporter.logging](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.logging)
- [otelcol.exporter.loki](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.loki)
- [otelcol.exporter.otlp](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.otlp)
- [otelcol.exporter.otlphttp](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.otlphttp)
- [otelcol.exporter.prometheus](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.exporter.prometheus)
- [otelcol.processor.attributes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.attributes)
- [otelcol.processor.batch](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.batch)
- [otelcol.processor.discovery](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.discovery)
- [otelcol.processor.filter](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.filter)
- [otelcol.processor.k8sattributes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.k8sattributes)
- [otelcol.processor.memory_limiter](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.memory_limiter)
- [otelcol.processor.probabilistic_sampler](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.probabilistic_sampler)
- [otelcol.processor.resourcedetection](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.resourcedetection)
- [otelcol.processor.span](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.span)
- [otelcol.processor.tail_sampling](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.tail_sampling)
- [otelcol.processor.transform](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.transform)
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF OpenTelemetry `otelcol.Consumer` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### OpenTelemetry `otelcol.Consumer` Consumers
The following components, grouped by namespace, _consume_ OpenTelemetry `otelcol.Consumer`.

<!-- START GENERATED SECTION: CONSUMERS OF OpenTelemetry `otelcol.Consumer` -->

{{< collapse title="faro" >}}
- [faro.receiver](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/faro.receiver)
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.connector.host_info](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.host_info)
- [otelcol.connector.servicegraph](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.servicegraph)
- [otelcol.connector.spanlogs](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.spanlogs)
- [otelcol.connector.spanmetrics](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.connector.spanmetrics)
- [otelcol.processor.attributes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.attributes)
- [otelcol.processor.batch](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.batch)
- [otelcol.processor.discovery](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.discovery)
- [otelcol.processor.filter](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.filter)
- [otelcol.processor.k8sattributes](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.k8sattributes)
- [otelcol.processor.memory_limiter](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.memory_limiter)
- [otelcol.processor.probabilistic_sampler](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.probabilistic_sampler)
- [otelcol.processor.resourcedetection](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.resourcedetection)
- [otelcol.processor.span](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.span)
- [otelcol.processor.tail_sampling](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.tail_sampling)
- [otelcol.processor.transform](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.processor.transform)
- [otelcol.receiver.jaeger](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.jaeger)
- [otelcol.receiver.kafka](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.kafka)
- [otelcol.receiver.loki](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.loki)
- [otelcol.receiver.opencensus](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.opencensus)
- [otelcol.receiver.otlp](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.otlp)
- [otelcol.receiver.prometheus](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.prometheus)
- [otelcol.receiver.vcenter](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.vcenter)
- [otelcol.receiver.zipkin](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/otelcol.receiver.zipkin)
{{< /collapse >}}

<!-- END GENERATED SECTION: CONSUMERS OF OpenTelemetry `otelcol.Consumer` -->

## Pyroscope `ProfilesReceiver`

The Pyroscope profiles are sent between components using `ProfilesReceiver`s.
`ProfilesReceiver`s are [capsules]({{< relref "../../concepts/config-language/expressions/types_and_values/#capsules" >}})
that are exported by components that can receive Pyroscope profiles. Components that
can consume Pyroscope profiles can be passed the `ProfilesReceiver` as an argument. Use the
following components to build your Pyroscope profiles pipeline:

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Pyroscope `ProfilesReceiver` Exporters
The following components, grouped by namespace, _export_ Pyroscope `ProfilesReceiver`.

<!-- START GENERATED SECTION: EXPORTERS OF Pyroscope `ProfilesReceiver` -->

{{< collapse title="pyroscope" >}}
- [pyroscope.write](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.write)
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Pyroscope `ProfilesReceiver` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Pyroscope `ProfilesReceiver` Consumers
The following components, grouped by namespace, _consume_ Pyroscope `ProfilesReceiver`.

<!-- START GENERATED SECTION: CONSUMERS OF Pyroscope `ProfilesReceiver` -->

{{< collapse title="pyroscope" >}}
- [pyroscope.ebpf](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.ebpf)
- [pyroscope.java](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.java)
- [pyroscope.scrape](https://grafana.com/docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.scrape)
{{< /collapse >}}

<!-- END GENERATED SECTION: CONSUMERS OF Pyroscope `ProfilesReceiver` -->
