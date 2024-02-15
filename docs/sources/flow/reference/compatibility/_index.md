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
to work correctly. It is recommended to always check component references for
details when working with Targets.

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Targets Exporters
The following components, grouped by namespace, _export_ Targets.

<!-- START GENERATED SECTION: EXPORTERS OF Targets -->

{{< collapse title="discovery" >}}
- [discovery.azure]({{< relref "../components/discovery.azure.md" >}})
- [discovery.consul]({{< relref "../components/discovery.consul.md" >}})
- [discovery.consulagent]({{< relref "../components/discovery.consulagent.md" >}})
- [discovery.digitalocean]({{< relref "../components/discovery.digitalocean.md" >}})
- [discovery.dns]({{< relref "../components/discovery.dns.md" >}})
- [discovery.docker]({{< relref "../components/discovery.docker.md" >}})
- [discovery.dockerswarm]({{< relref "../components/discovery.dockerswarm.md" >}})
- [discovery.ec2]({{< relref "../components/discovery.ec2.md" >}})
- [discovery.eureka]({{< relref "../components/discovery.eureka.md" >}})
- [discovery.file]({{< relref "../components/discovery.file.md" >}})
- [discovery.gce]({{< relref "../components/discovery.gce.md" >}})
- [discovery.hetzner]({{< relref "../components/discovery.hetzner.md" >}})
- [discovery.http]({{< relref "../components/discovery.http.md" >}})
- [discovery.ionos]({{< relref "../components/discovery.ionos.md" >}})
- [discovery.kubelet]({{< relref "../components/discovery.kubelet.md" >}})
- [discovery.kubernetes]({{< relref "../components/discovery.kubernetes.md" >}})
- [discovery.kuma]({{< relref "../components/discovery.kuma.md" >}})
- [discovery.lightsail]({{< relref "../components/discovery.lightsail.md" >}})
- [discovery.linode]({{< relref "../components/discovery.linode.md" >}})
- [discovery.marathon]({{< relref "../components/discovery.marathon.md" >}})
- [discovery.nerve]({{< relref "../components/discovery.nerve.md" >}})
- [discovery.nomad]({{< relref "../components/discovery.nomad.md" >}})
- [discovery.openstack]({{< relref "../components/discovery.openstack.md" >}})
- [discovery.ovhcloud]({{< relref "../components/discovery.ovhcloud.md" >}})
- [discovery.process]({{< relref "../components/discovery.process.md" >}})
- [discovery.puppetdb]({{< relref "../components/discovery.puppetdb.md" >}})
- [discovery.relabel]({{< relref "../components/discovery.relabel.md" >}})
- [discovery.scaleway]({{< relref "../components/discovery.scaleway.md" >}})
- [discovery.serverset]({{< relref "../components/discovery.serverset.md" >}})
- [discovery.triton]({{< relref "../components/discovery.triton.md" >}})
- [discovery.uyuni]({{< relref "../components/discovery.uyuni.md" >}})
{{< /collapse >}}

{{< collapse title="local" >}}
- [local.file_match]({{< relref "../components/local.file_match.md" >}})
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.exporter.apache]({{< relref "../components/prometheus.exporter.apache.md" >}})
- [prometheus.exporter.azure]({{< relref "../components/prometheus.exporter.azure.md" >}})
- [prometheus.exporter.blackbox]({{< relref "../components/prometheus.exporter.blackbox.md" >}})
- [prometheus.exporter.cadvisor]({{< relref "../components/prometheus.exporter.cadvisor.md" >}})
- [prometheus.exporter.cloudwatch]({{< relref "../components/prometheus.exporter.cloudwatch.md" >}})
- [prometheus.exporter.consul]({{< relref "../components/prometheus.exporter.consul.md" >}})
- [prometheus.exporter.dnsmasq]({{< relref "../components/prometheus.exporter.dnsmasq.md" >}})
- [prometheus.exporter.elasticsearch]({{< relref "../components/prometheus.exporter.elasticsearch.md" >}})
- [prometheus.exporter.gcp]({{< relref "../components/prometheus.exporter.gcp.md" >}})
- [prometheus.exporter.github]({{< relref "../components/prometheus.exporter.github.md" >}})
- [prometheus.exporter.kafka]({{< relref "../components/prometheus.exporter.kafka.md" >}})
- [prometheus.exporter.memcached]({{< relref "../components/prometheus.exporter.memcached.md" >}})
- [prometheus.exporter.mongodb]({{< relref "../components/prometheus.exporter.mongodb.md" >}})
- [prometheus.exporter.mssql]({{< relref "../components/prometheus.exporter.mssql.md" >}})
- [prometheus.exporter.mysql]({{< relref "../components/prometheus.exporter.mysql.md" >}})
- [prometheus.exporter.oracledb]({{< relref "../components/prometheus.exporter.oracledb.md" >}})
- [prometheus.exporter.postgres]({{< relref "../components/prometheus.exporter.postgres.md" >}})
- [prometheus.exporter.process]({{< relref "../components/prometheus.exporter.process.md" >}})
- [prometheus.exporter.redis]({{< relref "../components/prometheus.exporter.redis.md" >}})
- [prometheus.exporter.self]({{< relref "../components/prometheus.exporter.self.md" >}})
- [prometheus.exporter.snmp]({{< relref "../components/prometheus.exporter.snmp.md" >}})
- [prometheus.exporter.snowflake]({{< relref "../components/prometheus.exporter.snowflake.md" >}})
- [prometheus.exporter.squid]({{< relref "../components/prometheus.exporter.squid.md" >}})
- [prometheus.exporter.statsd]({{< relref "../components/prometheus.exporter.statsd.md" >}})
- [prometheus.exporter.unix]({{< relref "../components/prometheus.exporter.unix.md" >}})
- [prometheus.exporter.vsphere]({{< relref "../components/prometheus.exporter.vsphere.md" >}})
- [prometheus.exporter.windows]({{< relref "../components/prometheus.exporter.windows.md" >}})
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Targets -->


<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Targets Consumers
The following components, grouped by namespace, _consume_ Targets.

<!-- START GENERATED SECTION: CONSUMERS OF Targets -->

{{< collapse title="discovery" >}}
- [discovery.process]({{< relref "../components/discovery.process.md" >}})
- [discovery.relabel]({{< relref "../components/discovery.relabel.md" >}})
{{< /collapse >}}

{{< collapse title="local" >}}
- [local.file_match]({{< relref "../components/local.file_match.md" >}})
{{< /collapse >}}

{{< collapse title="loki" >}}
- [loki.source.docker]({{< relref "../components/loki.source.docker.md" >}})
- [loki.source.file]({{< relref "../components/loki.source.file.md" >}})
- [loki.source.kubernetes]({{< relref "../components/loki.source.kubernetes.md" >}})
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.processor.discovery]({{< relref "../components/otelcol.processor.discovery.md" >}})
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.scrape]({{< relref "../components/prometheus.scrape.md" >}})
{{< /collapse >}}

{{< collapse title="pyroscope" >}}
- [pyroscope.ebpf]({{< relref "../components/pyroscope.ebpf.md" >}})
- [pyroscope.java]({{< relref "../components/pyroscope.java.md" >}})
- [pyroscope.scrape]({{< relref "../components/pyroscope.scrape.md" >}})
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
- [otelcol.receiver.prometheus]({{< relref "../components/otelcol.receiver.prometheus.md" >}})
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.relabel]({{< relref "../components/prometheus.relabel.md" >}})
- [prometheus.remote_write]({{< relref "../components/prometheus.remote_write.md" >}})
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Prometheus `MetricsReceiver` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Prometheus `MetricsReceiver` Consumers
The following components, grouped by namespace, _consume_ Prometheus `MetricsReceiver`.

<!-- START GENERATED SECTION: CONSUMERS OF Prometheus `MetricsReceiver` -->

{{< collapse title="otelcol" >}}
- [otelcol.exporter.prometheus]({{< relref "../components/otelcol.exporter.prometheus.md" >}})
{{< /collapse >}}

{{< collapse title="prometheus" >}}
- [prometheus.operator.podmonitors]({{< relref "../components/prometheus.operator.podmonitors.md" >}})
- [prometheus.operator.probes]({{< relref "../components/prometheus.operator.probes.md" >}})
- [prometheus.operator.servicemonitors]({{< relref "../components/prometheus.operator.servicemonitors.md" >}})
- [prometheus.receive_http]({{< relref "../components/prometheus.receive_http.md" >}})
- [prometheus.relabel]({{< relref "../components/prometheus.relabel.md" >}})
- [prometheus.scrape]({{< relref "../components/prometheus.scrape.md" >}})
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
- [loki.echo]({{< relref "../components/loki.echo.md" >}})
- [loki.process]({{< relref "../components/loki.process.md" >}})
- [loki.relabel]({{< relref "../components/loki.relabel.md" >}})
- [loki.write]({{< relref "../components/loki.write.md" >}})
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.receiver.loki]({{< relref "../components/otelcol.receiver.loki.md" >}})
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Loki `LogsReceiver` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Loki `LogsReceiver` Consumers
The following components, grouped by namespace, _consume_ Loki `LogsReceiver`.

<!-- START GENERATED SECTION: CONSUMERS OF Loki `LogsReceiver` -->

{{< collapse title="faro" >}}
- [faro.receiver]({{< relref "../components/faro.receiver.md" >}})
{{< /collapse >}}

{{< collapse title="loki" >}}
- [loki.process]({{< relref "../components/loki.process.md" >}})
- [loki.relabel]({{< relref "../components/loki.relabel.md" >}})
- [loki.source.api]({{< relref "../components/loki.source.api.md" >}})
- [loki.source.awsfirehose]({{< relref "../components/loki.source.awsfirehose.md" >}})
- [loki.source.azure_event_hubs]({{< relref "../components/loki.source.azure_event_hubs.md" >}})
- [loki.source.cloudflare]({{< relref "../components/loki.source.cloudflare.md" >}})
- [loki.source.docker]({{< relref "../components/loki.source.docker.md" >}})
- [loki.source.file]({{< relref "../components/loki.source.file.md" >}})
- [loki.source.gcplog]({{< relref "../components/loki.source.gcplog.md" >}})
- [loki.source.gelf]({{< relref "../components/loki.source.gelf.md" >}})
- [loki.source.heroku]({{< relref "../components/loki.source.heroku.md" >}})
- [loki.source.journal]({{< relref "../components/loki.source.journal.md" >}})
- [loki.source.kafka]({{< relref "../components/loki.source.kafka.md" >}})
- [loki.source.kubernetes]({{< relref "../components/loki.source.kubernetes.md" >}})
- [loki.source.kubernetes_events]({{< relref "../components/loki.source.kubernetes_events.md" >}})
- [loki.source.podlogs]({{< relref "../components/loki.source.podlogs.md" >}})
- [loki.source.syslog]({{< relref "../components/loki.source.syslog.md" >}})
- [loki.source.windowsevent]({{< relref "../components/loki.source.windowsevent.md" >}})
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.exporter.loki]({{< relref "../components/otelcol.exporter.loki.md" >}})
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
- [otelcol.connector.servicegraph]({{< relref "../components/otelcol.connector.servicegraph.md" >}})
- [otelcol.connector.spanlogs]({{< relref "../components/otelcol.connector.spanlogs.md" >}})
- [otelcol.connector.spanmetrics]({{< relref "../components/otelcol.connector.spanmetrics.md" >}})
- [otelcol.exporter.loadbalancing]({{< relref "../components/otelcol.exporter.loadbalancing.md" >}})
- [otelcol.exporter.logging]({{< relref "../components/otelcol.exporter.logging.md" >}})
- [otelcol.exporter.loki]({{< relref "../components/otelcol.exporter.loki.md" >}})
- [otelcol.exporter.otlp]({{< relref "../components/otelcol.exporter.otlp.md" >}})
- [otelcol.exporter.otlphttp]({{< relref "../components/otelcol.exporter.otlphttp.md" >}})
- [otelcol.exporter.prometheus]({{< relref "../components/otelcol.exporter.prometheus.md" >}})
- [otelcol.processor.attributes]({{< relref "../components/otelcol.processor.attributes.md" >}})
- [otelcol.processor.batch]({{< relref "../components/otelcol.processor.batch.md" >}})
- [otelcol.processor.discovery]({{< relref "../components/otelcol.processor.discovery.md" >}})
- [otelcol.processor.filter]({{< relref "../components/otelcol.processor.filter.md" >}})
- [otelcol.processor.k8sattributes]({{< relref "../components/otelcol.processor.k8sattributes.md" >}})
- [otelcol.processor.memory_limiter]({{< relref "../components/otelcol.processor.memory_limiter.md" >}})
- [otelcol.processor.probabilistic_sampler]({{< relref "../components/otelcol.processor.probabilistic_sampler.md" >}})
- [otelcol.processor.resourcedetection]({{< relref "../components/otelcol.processor.resourcedetection.md" >}})
- [otelcol.processor.span]({{< relref "../components/otelcol.processor.span.md" >}})
- [otelcol.processor.tail_sampling]({{< relref "../components/otelcol.processor.tail_sampling.md" >}})
- [otelcol.processor.transform]({{< relref "../components/otelcol.processor.transform.md" >}})
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF OpenTelemetry `otelcol.Consumer` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### OpenTelemetry `otelcol.Consumer` Consumers
The following components, grouped by namespace, _consume_ OpenTelemetry `otelcol.Consumer`.

<!-- START GENERATED SECTION: CONSUMERS OF OpenTelemetry `otelcol.Consumer` -->

{{< collapse title="faro" >}}
- [faro.receiver]({{< relref "../components/faro.receiver.md" >}})
{{< /collapse >}}

{{< collapse title="otelcol" >}}
- [otelcol.connector.servicegraph]({{< relref "../components/otelcol.connector.servicegraph.md" >}})
- [otelcol.connector.spanlogs]({{< relref "../components/otelcol.connector.spanlogs.md" >}})
- [otelcol.connector.spanmetrics]({{< relref "../components/otelcol.connector.spanmetrics.md" >}})
- [otelcol.processor.attributes]({{< relref "../components/otelcol.processor.attributes.md" >}})
- [otelcol.processor.batch]({{< relref "../components/otelcol.processor.batch.md" >}})
- [otelcol.processor.discovery]({{< relref "../components/otelcol.processor.discovery.md" >}})
- [otelcol.processor.filter]({{< relref "../components/otelcol.processor.filter.md" >}})
- [otelcol.processor.k8sattributes]({{< relref "../components/otelcol.processor.k8sattributes.md" >}})
- [otelcol.processor.memory_limiter]({{< relref "../components/otelcol.processor.memory_limiter.md" >}})
- [otelcol.processor.probabilistic_sampler]({{< relref "../components/otelcol.processor.probabilistic_sampler.md" >}})
- [otelcol.processor.resourcedetection]({{< relref "../components/otelcol.processor.resourcedetection.md" >}})
- [otelcol.processor.span]({{< relref "../components/otelcol.processor.span.md" >}})
- [otelcol.processor.tail_sampling]({{< relref "../components/otelcol.processor.tail_sampling.md" >}})
- [otelcol.processor.transform]({{< relref "../components/otelcol.processor.transform.md" >}})
- [otelcol.receiver.jaeger]({{< relref "../components/otelcol.receiver.jaeger.md" >}})
- [otelcol.receiver.kafka]({{< relref "../components/otelcol.receiver.kafka.md" >}})
- [otelcol.receiver.loki]({{< relref "../components/otelcol.receiver.loki.md" >}})
- [otelcol.receiver.opencensus]({{< relref "../components/otelcol.receiver.opencensus.md" >}})
- [otelcol.receiver.otlp]({{< relref "../components/otelcol.receiver.otlp.md" >}})
- [otelcol.receiver.prometheus]({{< relref "../components/otelcol.receiver.prometheus.md" >}})
- [otelcol.receiver.vcenter]({{< relref "../components/otelcol.receiver.vcenter.md" >}})
- [otelcol.receiver.zipkin]({{< relref "../components/otelcol.receiver.zipkin.md" >}})
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
- [pyroscope.write]({{< relref "../components/pyroscope.write.md" >}})
{{< /collapse >}}

<!-- END GENERATED SECTION: EXPORTERS OF Pyroscope `ProfilesReceiver` -->

<!-- NOTE: this title is used as an anchor in links. Do not change. -->
### Pyroscope `ProfilesReceiver` Consumers
The following components, grouped by namespace, _consume_ Pyroscope `ProfilesReceiver`.

<!-- START GENERATED SECTION: CONSUMERS OF Pyroscope `ProfilesReceiver` -->

{{< collapse title="pyroscope" >}}
- [pyroscope.ebpf]({{< relref "../components/pyroscope.ebpf.md" >}})
- [pyroscope.java]({{< relref "../components/pyroscope.java.md" >}})
- [pyroscope.scrape]({{< relref "../components/pyroscope.scrape.md" >}})
{{< /collapse >}}

<!-- END GENERATED SECTION: CONSUMERS OF Pyroscope `ProfilesReceiver` -->
