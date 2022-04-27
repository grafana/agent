# Depricating agent-operator

* Date: 2022-03-28
* Author: Craig Peterson (@captncraig)
* PR:
* Status: Draft

## Background

The Grafana Agent Operator was created as a convenience for deploying Grafana Agent to Kubernetes clusters. 
It does so by introducing custom resources to represent Grafana Agent instances, metrics and logging backends, as well as
scrape configs. One problem that comes up consistently is that these CRDs are yet another way to represent the Grafana Agent config. The operator
requires extra maintenance to make new agent features availible to users, and it can get out of sync easily. Without being well understood,
the operator model can lead to confusion for new users, and make debugging the agent more difficult.

I propose we deprecate the operator, and instead focus on making agent deployments as seamless as possible for users to configure and deploy it directly.

### Purpose of the agent

The agent-operator is designed to be the easiest way to deploy Grafana Agent to kubernetes. It watches two main sets of CRDs:

- Prometheus operator crds for podmonitors, servicemoniters, etc.
- Grafana Agent-specific CRDs for declaring agent instances, metrics endpoints and so forth.

It allows users with existing prometheus-operator infrastructure to reuse those monitor resources without needing to redefine scrape configs for the agent.

### Problems with the agent

- The agent specific CRDs are similar to, but distinct from the internal agent config types.
- Operator config is not always up to date with agent config. Major agent features (tracing for example), take additional work to port into the operator crds.
- Kubernetes operators are less understood by the larger agent squad. Maintenance on the operator is slower than on the agent itself.
- Documentation must take into account that there are two fully supported ways to deploy and configure the agent. Operator crds are a completely different format than the agent config yaml.

### Alternatives

The above lead me to wonder if the operator is worth the effort of maintaining as a seperate app. If our goal is to make deployment of the agent to kubernetes as easy as possible, perhaps we would be better served by providing a happy path to deploying customized manifests directly. 

The first large step we should take is publishing an official helm chart. It should be usable to deploy the agent with a working config extremely easily, while still allowing fully custom agent config.

The biggest missing features in deprecating the operator would probably be native support for the prometheus operator Monitor crds. We should explore the possibility of having the agent itself watch those types and merge into its own config, or some other method for supporting those (if they are important to our user base).