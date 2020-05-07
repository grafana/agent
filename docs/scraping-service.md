# Scraping Service Mode

Scraping Service Mode is a third operational mode of the Grafana Cloud Agent
that allows for clustering a set of Agent processes and distributing scrape load
between them.

Determining what to scrape is done by storing instance configuration files to an
[API](./api.md) which then stores the configuration files in a KV store backend.
All agents in the cluster must use the same KV store so they read the same set
of config files.

The "instance" refers to the process of:

- Service discovery for all `scrape_configs` within that loaded config
- Scrapes metrics from all discovered targets
- Stores data in its own Write-Ahead Log specific to the loaded config
- Remote Writes scraped metrics to the configured `remote_write` clients
  specified within the loaded config.


The "instance configuration file," then, is the configuration file that
specifies the set of `scrape_configs` and `remote_write` endpoints. Each process
of the Grafana Cloud Agent can be running multiple independent instances at
once.

When Scraping Service Mode is enabled, Agents disallow specifying
instance configurations locally in the configuration file; using the KV store
is required.

GitOps-friendly tooling is planned to automatically load instance configuration
files

## Distributed Hash Ring

Scraping Service Mode uses a Distributed Hash Ring (commonly just called "the
ring") to cluster agents and to shard configurations within that ring. Each
Agent joins the ring with a random distinct set of _tokens_ that are used for
sharding. The default number of generated tokens is 128.

The Distributed Hash Ring is also stored in a KV store. Since a KV store is
also needed for storing configuration files, it is common practice to re-use
that same KV store for the ring.

When sharding, the Agent currently uses the entire contents of a config file
stored in the KV store for load distribution. The hash of the config file is
used as the _key_ in the ring and looks up an Agent that should be responsible
for that config. "Price is Right" rules are used for the Agent lookup; the Agent
owning the token with the closest value to the key without going over is
responsible for the config.

When an Agent receives a new config that it is responsible for, it launches a
new instance from the instance config. If a config is deleted from the KV store,
this will be detected by the owning Agent and it will stop the metric collection
process for that config file.

Scraping Service Mode currently does not support replication; only one agent
at a time will be responsible for scraping a certain config.

### Resharding

When a new Agent joins or leaves the cluster, the set of tokens in the ring may
cause configurations to hash to a new Agent. The process of responding to this
action is called "resharding."

Resharding is run:

1. When an Agent joins the ring
2. When an Agent leaves the ring
3. When the KV store sends a notification indicating a config has changed.
4. On a specified interval in case KV change events have not fired.

The resharding process involves each Agent retrieving the full set of
configurations stored in the KV store and determining if:

1. The config owned by the current resharding Agent has changed and needs to
   be reloaded.
2. The config is no longer owned by the current resharding Agent and the
   associated instance should be stopped.
3. The config has been deleted and the associated instance should be stopped.

## Best Practices

Because distribution is determined by the number of config files and not how
many targets exist per config file, the best amount of distribution is achieved
when each config file has the lowest amount of targets possible. The best
distribution will be achieved if each config file stored in the KV store is
limited to one static config with only one target.

A better distribution mechanism that distributes based on discovered targets is
planned for the future.

## Example

Here's an example `agent.yaml` config file that uses the same `etcd` server for
both configuration storage and the distributed hash ring storage:

```yaml
server:
  log_level: debug
  http_listen_port: 12345

prometheus:
  global:
    scrape_interval: 5s
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

Note that there are no instance configs present in this example; instance
configs must be passed to the API for the Agent to start scraping metrics.
See [the docker-compose Scraping Service Example](../example/README.md)
for how to run a Scraping Service Agent cluster locally.
