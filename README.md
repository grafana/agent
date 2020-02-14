# Grafana Cloud Agent

## Running Example

A docker-compose config is provided in `example/`. Before you can run it, you
need to build the agent image:

```
make agent-image
cd example/ && docker-compose up -d
```

This will create Cortex storage in `/tmp/cortex` and the agent WAL in
`/tmp/agent`; you may want to delete those directories after you are done
testing.

Once the containers are running, a Grafana instance will be exposed at
`http://localhost:3000` with Cortex as the only datasource. You should shortly
see metrics in Cortex that are sent from the agent.

Slightly modified versions of the Prometheus mixin dashboards will be added to
the launched Grafana instance; see the "Agent" and "Agent Remote Write"
dashboards for details.

The agent will be exposed locally at `http://localhost:12345`; this is useful
for running pprof against:

```
go tool pprof -http=:6060 http://localhost:12345/debug/pprof/heap?debug=1`
```

Useful queries to run once everything is running:

1. `agent_wal_storage_active_series`: How many series are active in the WAL
2. `cortex_ingester_memory_series`: How many series are active in Cortex.
   Should be equal to the previous metric.
3. `go_memstats_heap_inuse_bytes{container="agent"} / 1e6`: Current memory
   usage of agent in megabytes.
4. `max by (container,instance,job)
   (avg_over_time(go_memstats_heap_inuse_bytes[10m])) / 1e6`: Current memory
   usage of the agent and Cortex averaged out from the last 10 minutes.

