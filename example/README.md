# Example

This directory contains an example deployment of the Grafana Cloud Agent with
the following components:

1. Cortex to store metrics
2. Grafana Cloud Agent to collect metrics
3. Grafana to visualize metrics
4. Avalanche to load test the Agent.

To get started, run the following from this directory:

```
docker-compose up -d
```

This will create Cortex storage in `/tmp/cortex` and the agent WAL in
`/tmp/agent`; you may want to delete those directories after you are done
testing.

Once the containers are running, a Grafana instance will be exposed at
`http://localhost:3000` with Cortex as the only datasource. You should shortly
see metrics in Cortex that are sent from the agent. Agent operational dashboards
are included in the deployment.

The Agent is exposed on the host at `http://localhost:12345`.

## Hacking on the Example

The reduced memory requirements is a critical feature of the Agent, and
the example provides a good launching point to end-to-end test and validate
the usage.

To get a memory profile, you can use `pprof` against the Agent:

```
go tool pprof -http=:6060 http://localhost:12345/debug/pprof/heap?debug=1
```

Useful one-off queries to run once everything is up:

1. `agent_wal_storage_active_series`: How many series are active in the WAL
2. `cortex_ingester_memory_series`: How many series are active in Cortex.
   Should be equal to the previous metric.
3. `go_memstats_heap_inuse_bytes{container="agent"} / 1e6`: Current memory
   usage of agent in megabytes.
4. `max by (container,instance,job)
   (avg_over_time(go_memstats_heap_inuse_bytes[10m])) / 1e6`: Current memory
   usage of the agent and Cortex averaged out from the last 10 minutes.

