---
title: Run Grafana Agent Flow
weight: 100
---

# Run Grafana Agent Flow

This topic describes how to run Grafana Agent Flow.

## Before you begin

* [Install Grafana Agent][]

[Install Grafana Agent]: {{< relref "../../set-up" >}}

## Steps

To run Grafana Agent Flow, follow these steps:

1. Set the `AGENT_MODE` environment variable on your system to `flow`.

2. Create a configuration file for Grafana Agent Flow to use, saving it to your
   system with a `.river` file extension. This file can be used as a starting
   point:

   ```river
   logging {
     level  = "info"
     format = "logfmt"
   }
   ```

3. Run the following command in your terminal, replacing `FILE_PATH` with the
   path to the configuration file you created:

   ```bash
   grafana-agent run FILE_PATH
   ```

4. To confirm Grafana Agent Flow is running, visit <http://localhost:12345> in a
   web browser to view the [Grafana Agent Flow UI][ui].

For more information on the `grafana-agent run` command, see the [`agent run`
command]({{< relref "../reference/cli/run.md" >}}).

[ui]: {{< relref "../monitoring/debugging.md#grafana-agent-flow-ui" >}}
