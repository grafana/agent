---
aliases:
- /docs/grafana-cloud/agent/flow/tasks/send-profiles-to-pyroscope/
canonical: https://grafana.com/docs/agent/latest/flow/send-profiles-to-pyroscope/
description: Send profiles to Pyroscope 
title: Send profiles to Pyroscope 
weight: 120
---

# Send profiles to Pyroscope

Learn how to configure {{< param "PRODUCT_NAME" >}} to collect profiles and forward them to a [Pyroscope Server][].

This topic describes how to:

* Setup [Pyroscope Server][].
* Setup {{< param "PRODUCT_NAME" >}} to collect profiles and send them to [Pyroscope Server][].

## Components used in this topic

* [pyroscope.write][]
* [pyroscope.scrape][]initial

## Before you begin

* Install [docker][].

## Steps

1. Setup Pyroscope
     
   The easiest path to setting up Pyroscope is using `docker run -it -p 4040:4040 grafana/pyroscope`. This will spin up a Docker instance running Pyroscope on port `4040`. The server will autoscrape itself, and after a few minutes, data will appear. If you immediately access the server, it may error with `No applications available` until it has scraped itself.  
   
   ![Initial Pyroscope server screen](/media/oss/agent/initial-pyro.png)

2. Install the [latest][] version of {{< param "PRODUCT_NAME" >}} for your operating system.

3. Add the below configuration to `agent.river` file in the same directory as the file downloaded above.

   ```river
   pyroscope.scrape "agent" {
           targets    = [{"__address__" = "localhost:12345", "service_name" = "agent"}]
           forward_to = [pyroscope.write.local.receiver]
   
           profiling_config {
                   profile.process_cpu {
                           enabled = true
                   }
   
                   profile.godeltaprof_memory {
                           enabled = true
                   }
          }
   }

   pyroscope.write "local" {
           endpoint {
                   url = "http://localhost:4040"
           }
   }
    
   ```

   This configuration will scrape the `localhost:12345/-/pprof` endpoint for CPU and memory data every 60 seconds and send those profiles to the Pyroscope server. The `localhost:12345` endpoint is the default host and port for {{< param "PRODUCT_NAME" >}}.

4. Run {{< param "PRODUCT_NAME" >}} with `AGENT_MODE=flow ./grafana-agent-linux-amd64 run ./agent.river`. The exact executable name will change depending on the platform. Wait 2 minutes, this will give time for startup and a scrape to occur.

5. Open `http://localhost:4040` in a web browser.

6. Select `agent` from the dropdown. This name is derived from `service_name` specified in `agent.river`.

   ![Select agent from dropdown](/media/oss/agent/select-pyro.png)

7. Select any CPU or metric you want to view.

   ![Agent CPU](/media/oss/agent/normal-pyro.png)

## Using Pyroscope with Grafana Cloud

1. Login to your account at [Grafana Cloud][]- [ ] Tests updated
- [ ] Config converters updated](https://docs.docker.com/engine/install/)
2. Go to your stack and select Details.
3. Select Details under the Pyroscope logo.
4. Generate an API key.
5. Update the `agent.river` file to look like the following configuration. Fill in `URL`,`USERNAME`, and `PASSWORD` with the information from the Pyroscope details page.

   ```river
   pyroscope.scrape "agent" {
           t- [ ] Tests updated
- [ ] Config converters updated](https://docs.docker.com/engine/install/)argets    = [{"__address__" = "localhost:12345", "service_name" = "agent"}]
           forward_to = [pyroscope.write.local.receiver]
   
           profiling_config {
                   profile.process_cpu {
                           enabled = true
                   }
   
                   profile.godeltaprof_memory {
                           enabled = true
                   }
          }
   }
   
   pyroscope.write "local" {
           endpoint {
                   url = URL
                   basic_auth {
                           username = USERNAME
                           password = PASSWORD
                   }
           }
   }
   ```



## Additional links

* [Set up Go profiling in pull mode][]

[latest]: https://github.com/grafana/agent/releases/latest
[Set up Go profiling in pull mode]: https://grafana.com/docs/pyroscope/v1.2.x/configure-client/grafana-agent/go_pull/
[Pyroscope Server]: https://github.com/grafana/pyroscope#-quick-start-run-pyroscope-locally
[Grafana Cloud]: https://grafana.com/ 
[docker]: https://docs.docker.com/engine/install/

{{% docs/reference %}}  
[pyroscope.write]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.write.md"
[pyroscope.write]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/components/pyroscope.write.md"
[pyroscope.scrape]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/components/pyroscope.scrape.md"
[pyroscope.scrape]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/send-data/agent/flow/reference/components/pyroscope.scrape.md"
{{% /docs/reference %}}
