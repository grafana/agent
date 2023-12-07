# Example

This directory contains a Docker Compose 3 environment that can be used to test
Grafana Agent.

By default, the following services are exposed:

1. Mimir for storing metrics (localhost:9009)
2. Grafana for visualizing telemetry (localhost:3000)
3. Loki for storing logs (localhost:3100)
4. Tempo for storing traces (localhost:3200)
5. Avalanche for a sample /metrics endpoint to scrape (localhost:9001).

Run the following to bring up the environment:

```
docker-compose up -d
```

By default, the Docker Compose environment doesn't include a Grafana Agent
container. This lets you test the agent externally, especially useful when
validating code changes. You can enable the included Grafana Agent by passing
`agent` to the profiles list: `docker compose --profile=agent up -d`. When
running, the Agent exposes its HTTP endpoint at localhost:12345. This address
can be changed with the `--server.http.address` flag (e.g.,
`--server.http.address=127.0.0.1:8000`).

The Docker Compose environment heavily relies on profiles to enable optional
features. You can pass multiple profiles by passing the flag multiple times:
`docker compose --profile agent --profile integrations up -d`.

## Running Integrations

The Docker Compose environment includes example services to point integrations
at (e.g., mysql, redis, consul, etc.). You can run all integrations using the
`integrations` profile, or run individual services with a profile name matching
the integration (i.e., enabling the `dnsmasq_exporter` profile to specifically
enable the dnsmasq service).

Enabling specific integration profiles is useful when you only want to test a
single integration, as the `integrations` profile can be resource intensive.

## Visualizing

Grafana is exposed at `http://localhost:3000`, and includes some useful
dashboards:

* The `Agent` dashboard gives a very high-level overview of running agents.

* The `Agent Prometheus Remote Write` dashboard visualizes the current state of
  writing metrics to Mimir.

* The `Agent Tracing Pipeline` dashboard visualizes the current state of the
  tracing pipeline (if spans are being processed).

* The `Agent Operational` dashboard shows resource consumption of Grafana
  Agent. Not all panels will have data here, as they rely on metrics from other
  sources (i.e., cAdvisor).
