TODO(thampiotr): A doc that will describe how to collect and send metrics in a Kubernetes cluster.
This includes the installation steps via Helm, as a StatefulSet.

We describe adding persistent volume for the WAL and how to manage it.

This will use the [kubernetes module](https://github.com/grafana/agent-modules/tree/main/modules/kubernetes)
which uses annotation-driven approach. This doc will describe how to get a 
basic setup, e.g. scraping pods via adding `metrics.agent.grafana.com/scrape: "true"` annotation to a pod.

For more information, we'll refer to the [kubernetes module](https://github.com/grafana/agent-modules/tree/main/modules/kubernetes) docs.
We can also refer to clustering docs? Or maybe this doc should already cover this?
