---
aliases:
- /docs/grafana-cloud/agent/flow/tasks/sending-profiles-to-pyroscope/
canonical: https://grafana.com/docs/agent/latest/flow/sending-profiles-to-pyroscope/
description: Send profiles to Pyroscope 
title: Send profiles Pyroscope 
weight: 120
---

# Send profiles to Pyroscope

Learn how to configure {{< param "PRODUCT_NAME" >}} to collect [Pyroscope Server][] profiles and forward them to a [Pyroscope Server][].

This topic describes how to:

* Setup [Pyroscope Server][].
* Setup {{< param "PRODUCT_NAME" >}} to collect profiles and send them to [Pyroscope Server][].

## Components used in this topic

* [pyroscope.write][]
* [pyroscope.scrape][]

## Before you begin

* Download {{< param "PRODUCT_NAME" >}}.
* Install [docker][] if it is not already installed.

## Steps

1. Setup Pyroscope
     
   The easiest path to setting up Pyroscope is using `docker run -it -p 4040:4040 grafana/pyroscope`. This will spin up a docker instance running Pyroscope on port `4040`. The server will autoscrape itself and after a few minutes data will appear. If you immedietly access the server it may error with `No applications available` until it has scraped itself.  
   
   ![Initial Pyroscope server screen](../../../assets/tasks/initial-pyro.png)

2. Download the [latest][] version of {{< param "PRODUCT_NAME" >}} for the system your are running. 

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

   The above configuration will scrape the `localhost:12345/-/pprof` endpoint for cpu and memory evey 60 seconds and send those profiles to the Pyroscope server that was setup earlier. The `localhost:12345` endpoint is the default host and port for {{< param "PRODUCT_NAME" >}}.

4. Run {{< param "PRODUCT_NAME" >}} with `AGENT_MODE=flow ./grafana-agent-linux-amd64 run ./agent.river`. The exact executable name will change depending on the platform. Wait 2 minutes, this will give time for startup and a scrape to occur.

5. Open `http://localhost:4040` in a web browser.

6. Select `agent` from the dropdown. This name is derived from `service_name` specified in `agent.river`.

   ![Select agent from dropdown](../../../assets/tasks/select-pyro.png)

7. Select any cpu or metric you want to view.

   ![Agent CPU](../../../assets/tasks/normal-pyro.png)

## Using Pyroscope with Grafana Cloud

1. Login to your account at [Grafana Cloud][]
2. Go to your stack and select Details.
3. Select Details under the Pyroscope logo.
4. Generate a api key.
5. Update the `agent.river` file to look like the below configuration. Filling in `URL`,`USERNAME`, and `PASSWORD` with the information from the Pyroscope details page.

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
