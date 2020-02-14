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
