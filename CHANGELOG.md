# Next (master/unreleased)

- Fix various dashboard mixin problems from v0.1.0 (@rfratto)

# v0.1.0 (2020-03-16)

First (beta) release!

This release comes with support for scraping Prometheus metrics and
sharding the agent through the presence of a `host_filter` flag within the
Agent configuration file.

Note that enabling the `host_filter` flag currently works best when using our
preferred Kubernetes deployment, as it deploys the agent as a DaemonSet.
