---
aliases:
- ../../configuration/scraping-service/
- ../../scraping-service/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/scraping-service/
- /docs/grafana-cloud/send-data/agent/static/configuration/scraping-service/
canonical: https://grafana.com/docs/agent/latest/static/configuration/scraping-service/
description: Learn about the scraping service
menuTitle: Scraping service
title: Scraping service (Beta)
weight: 600
---

# Scraping service (Beta)

The Grafana Agent scraping service allows you to cluster a set of Agent processes and distribute the scrape load.

Determining what to scrape is done by writing instance configuration files to an
[API][api], which then stores the configuration files in a KV store backend.
All agents in the cluster **must** use the same KV store to see the same set
of configuration files.

Each process of the Grafana Agent can be running multiple independent
"instances" at once, where an "instance" refers to the combination of:

- Service discovery for all `scrape_configs` within that loaded configuration
- Scrapes metrics from all discovered targets
- Stores data in its own Write-Ahead Log specific to the loaded configuration
- Remote Writes scraped metrics to the configured `remote_write` destinations
  specified within the loaded configuration.

The "instance configuration file," then, is the configuration file that
specifies the set of `scrape_configs` and `remote_write` endpoints. For example,
a small instance configuration file looks like this:

```yaml
scrape_configs:
  - job_name: self-scrape
    static_configs:
      - targets: ['localhost:9090']
        labels:
          process: 'agent'
remote_write:
  - url: http://cortex:9009/api/prom/push
```

The full set of supported options for an instance configuration file is
available in the
[`metrics-config.md` file][metrics].

Multiple instance configuration files are necessary for sharding. Each
config file is distributed to a particular agent on the cluster based on the
hash of its contents.

When the scraping service is enabled, Agents **disallow** specifying
instance configurations locally in the configuration file; using the KV store
is required. [`agentctl`](#agentctl) can be used to manually sync
instance configuration files to the Agent's API server.

## Distributed hash ring

The scraping service uses a Distributed Hash Ring (commonly called "the
ring") to cluster agents and to shard configurations within that ring. Each
Agent joins the ring with a random distinct set of _tokens_ that are used for
sharding. The default number of generated tokens is 128.

The Distributed Hash Ring is also stored in a KV store. Since a KV store is
also needed for storing configuration files, it is encouraged to re-use
the same KV store for the ring.

When sharding, the Agent currently uses the name of a configuration file
stored in the KV store for load distribution. Configuration names are guaranteed to be
unique keys. The hash of the name is used as the _lookup key_ in the ring and
determines which agent (based on token) should be responsible for that configuration.
"Price is Right" rules are used for the Agent lookup; the Agent owning the token
with the closest value to the key without going over is responsible for the
configuration.

All Agents are simultaneously watching the KV store for changes to the set of
configuration files. When a configuration file is added or updated in the configuration
store, each Agent will run the configuration name hash through their copy of the Hash
Ring to determine if they are responsible for that config.

When an Agent receives a new configuration that it is responsible for, it launches a
new instance from the instance configuration. If a configuration is deleted from the KV store,
this will be detected by the owning Agent, and it will stop the metric collection
process for that configuration file.

When an Agent receives an event for an updated configuration file that they used to
be the owner of but are no longer the owner, the associated instance for that
configuration file is stopped for that Agent. This can happen when the cluster
size changes.

The scraping service currently does not support replication. Only one agent
at a time will be responsible for scraping a certain configuration.

### Resharding

When a new Agent joins or leaves the cluster, the set of tokens in the ring may
cause configurations to hash to a new Agent. The process of responding to this
action is called "resharding."

Resharding is run:

1. When an Agent joins the ring
2. When an Agent leaves the ring
3. When the KV store sends a notification indicating a configuration has changed.
4. On a specified interval if KV change events have not fired.

The resharding process involves each Agent retrieving the full set of
configurations stored in the KV store and determining if:

1. The configuration owned by the current resharding Agent has changed and needs to
   be reloaded.
2. The configuration is no longer owned by the current resharding Agent and the
   associated instance should be stopped.
3. The configuration has been deleted, and the associated instance should be stopped.

## Best practices

Because distribution is determined by the number of configuration files and not how
many targets exist per configuration file, the best amount of distribution is achieved
when each configuration file has the lowest amount of targets possible. The best
distribution will be achieved if each configuration file stored in the KV store is
limited to one static configuration with only one target.

## Example

Here's an example `agent.yaml` configuration file that uses the same `etcd` server for
both configuration storage and the distributed hash ring storage:

```yaml
server:
  log_level: debug

metrics:
  global:
    scrape_interval: 1m
  scraping_service:
    enabled: true
    kvstore:
      store: etcd
      etcd:
        endpoints:
          - etcd:2379
    lifecycler:
      ring:
        replication_factor: 1
        kvstore:
          store: etcd
          etcd:
            endpoints:
              - etcd:2379
```

Note that there are no instance configurations present in this example; instance
configurations must be passed to the API for the Agent to start scraping metrics.

## agentctl

`agentctl` is a tool included with this repository that helps users interact
with the new Config Management API. The `agentctl config-sync` subcommand uses
local YAML files as a source of truth and syncs their contents with the API.
Entries in the API not in the synced directory will be deleted.

`agentctl` is distributed in binary form with each release and as a Docker
container with the `grafana/agentctl` image. Tanka configurations that
utilize `grafana/agentctl` and sync a set of configurations to the API
are planned for the future.

## Debug Ring endpoint

You can use the `/debug/ring` endpoint to troubleshoot issues with the scraping service in Scraping Service Mode. 
It provides information about the Distributed Hash Ring and the current distribution of configurations among Agents in the cluster.
It also allows you to forget an instance in the ring manually.

You can access this endpoint by making an HTTP request to the Agent's API server.

Information returned by the `/debug/ring` endpoint includes:

- The list of Agents in the cluster, and their respective tokens used for sharding.
- The list of configuration files in the KV store and associated hash values used for lookup in the ring.
- The unique instance ID assigned to each instance of the Agent running in the cluster.
   The instance ID is a unique identifier assigned to each running instance of the Agent within the cluster.
   The exact details of the instance ID generation might be specific to the implementation of the Grafana Agent.
- The time of the "Last Heartbeat" of each instance. The Last Heartbeat is the last time the instance was active in the ring.

{{% docs/reference %}}
[api]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/api"
[api]: "/docs/grafana-cloud/ -> ../api"
[metrics]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/metrics-config"
[metrics]: "/docs/grafana-cloud/ -> ./metrics-config"
{{% /docs/reference %}}
