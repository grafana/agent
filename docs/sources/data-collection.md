---
aliases:
- ./data-collection/
- /docs/grafana-cloud/agent/data-collection/
- /docs/grafana-cloud/monitor-infrastructure/agent/data-collection/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/data-collection/
canonical: https://grafana.com/docs/agent/latest/data-collection/
menuTitle: Data collection
title: Grafana Agent data collection
description: Grafana Agent data collection
weight: 500
---

# Data collection

By default, Grafana Agent sends anonymous but uniquely identifiable usage information from
your Grafana Agent instance to Grafana Labs. These statistics are sent to `stats.grafana.org`.

Statistics help us better understand how Grafana Agent is used. This helps us prioritize features and documentation.

The usage information includes the following details:

* A randomly generated, anonymous unique ID (UUID).
* Timestamp of when the UID was first generated.
* Timestamp of when the report was created (by default, every four hours).
* Version of running Grafana Agent.
* Operating system Grafana Agent is running on.
* System architecture Grafana Agent is running on.
* List of enabled feature flags.
* List of enabled integrations.

This list may change over time. All newly reported data is documented in the CHANGELOG.

## Opt-out of data collection

You can use the `-disable-reporting` [command line flag][] to disable the reporting and opt-out of the data collection.

{{% docs/reference %}}
[command line flag]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/flow/reference/cli/run.md"
[command line flag]: "/docs/grafana-cloud/ -> /docs/grafana-cloud/flow/reference/cli/run.md"
{{% /docs/reference %}}