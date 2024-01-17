TODO(thampiotr): A doc that will describe how to collect and send logs in a Kubernetes cluster.
The prerequisite is to install as a DaemonSet. In the future we may point to API-based approach here.

We may describe mounting the pod logs volume and priviledged access.

This will use the [kubernetes module](https://github.com/grafana/agent-modules/tree/main/modules/kubernetes)
which uses annotation-driven approach. This doc will describe how to get a 
basic setup, e.g. collecting logs via adding `logs.agent.grafana.com/scrape: "true"` annotation to a pod.

For more information, we'll refer to the [kubernetes module](https://github.com/grafana/agent-modules/tree/main/modules/kubernetes) docs.
