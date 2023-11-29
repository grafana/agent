---
aliases:
- /docs/grafana-cloud/agent/flow/tutorials/flow-by-example/first-components-and-stdlib/full-config/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/tutorials/flow-by-example/first-components-and-stdlib/full-config/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/tutorials/flow-by-example/first-components-and-stdlib/full-config/
- /docs/grafana-cloud/send-data/agent/flow/tutorials/full-config/
canonical: https://grafana.com/docs/agent/latest/flow/tutorials/flow-by-example/first-components-and-stdlib/full-config/
description: Solution to the First Components and Standard Library Exercise
title: Solution
weight: 300
---

# Solution

[this section's example]: {{< relref "./guide/#exercise-for-the-reader" >}}

_This is the solution for [this section's example][]._

```river
// Configure your first components, learn about the standard library, and learn how to run the Agent!

// prometheus.exporter.redis collects information about Redis and exposes
// targets for other components to use
prometheus.exporter.redis "local_redis" {
    redis_addr = "localhost:6379"
}

prometheus.exporter.unix "localhost" { }

// prometheus.scrape scrapes the targets that it is configured with and forwards
// the metrics to other components (typically prometheus.relabel or prometheus.remote_write)
prometheus.scrape "default" {
    // This is scraping too often for typical use-cases, but is easier for testing and demo-ing!
    scrape_interval = "10s"

    // Here, prometheus.exporter.redis.local_redis.targets refers to the 'targets' export
    // of the prometheus.exporter.redis component with the label "local_redis".
    //
    // If you have more than one set of targets that you would like to scrape, you can use
    // the 'concat' function from the standard library to combine them.
    targets    = concat(prometheus.exporter.redis.local_redis.targets, prometheus.exporter.unix.localhost.targets)
    forward_to = [prometheus.remote_write.local_prom.receiver]
}

// prometheus.remote_write exports a 'receiver', which other components can forward
// metrics to and it will remote_write them to the configured endpoint(s)
prometheus.remote_write "local_prom" {
    endpoint {
        url = "http://localhost:9090/api/v1/write"
    }
}
```