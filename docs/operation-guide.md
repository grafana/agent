# Operation Guide 

## Prometheus "Instances"

The Grafana Cloud Agent defines a concept of a Prometheus _Instance_, which is
its own mini Prometheus-lite server. The Instance runs a combination of
Prometheus service discovery, scraping, a WAL for storage, and `remote_write`.

Instances allow for fine grained control of what data gets scraped and where it
gets sent. Users can easily define two Instances that scrape different subsets 
of metrics and send them two two completely different remote_write systems.

Instances are especially relevant to the [scraping service
mode](./scraping-service.md), where breaking up your scrape configs into
multiple Instances is required for sharding and balancing scrape load across a
cluster of Agents.

## Instance Sharing 

The v0.5.0 release of the Agent introduced the concept of _Instance sharing_,
which combines scrape_configs from compatible Instance configs into a single,
shared Instance. Instance configs are compatible when they have no differences
in configuration with the exception of what they scrape. `remote_write` configs
may also differ in the order which endpoints are declared, but the unsorted
`remote_writes` must still be an exact match.

The shared Instances mode is the new default, and the previous behavior is
deprecated. If you wish to restore the old behavior, set `instance_mode:
distinct` in the
[`prometheus_config`](./configuration-reference.md#prometheus_config) block of
your config file.

Shared Instances are completely transparent to the user with the exception of
exposed metrics. With `instance_mode: shared`, metrics for Prometheus components
(WAL, service discovery, remote_write, etc) have a `instance_group_name` label,
which is the hash of all settings used to determine the shared instance. When
`instance_mode: distinct` is set, the metrics for Prometheus components will
instead have an `instance_name` label, which matches the name set on the
individual Instance config. It is recommended to use the default of
`instance_mode: shared` unless you don't mind the performance hit and really
need granular metrics.

Users can use the [targets API](./api.md#list-current-scrape-targets) to see all
scraped targets, and the name of the shared instance they were assigned to.
