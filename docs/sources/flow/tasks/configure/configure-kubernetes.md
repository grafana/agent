---
aliases:
- /docs/grafana-cloud/agent/flow/tasks/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/tasks/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/tasks/configure/configure-kubernetes/
- /docs/grafana-cloud/send-data/agent/flow/tasks/configure/configure-kubernetes/
# Previous page aliases for backwards compatibility:
- /docs/grafana-cloud/agent/flow/setup/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/setup/configure/configure-kubernetes/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/setup/configure/configure-kubernetes/
- /docs/grafana-cloud/send-data/agent/flow/setup/configure/configure-kubernetes/
- ../../setup/configure/configure-kubernetes/ # /docs/agent/latest/flow/setup/configure/configure-kubernetes/
canonical: https://grafana.com/docs/agent/latest/flow/tasks/configure/configure-kubernetes/
description: Learn how to configure Grafana Agent Flow on Kubernetes
menuTitle: Kubernetes
title: Configure Grafana Agent Flow on Kubernetes
weight: 200
---

# Configure {{% param "PRODUCT_NAME" %}} on Kubernetes

TODO(thampiotr): The pre-requisite for this page is to have installed the Agent
on Kubernetes using a Helm chart, following one of the
tasks/kubernetes/collect-*.md pages. We'll describe here how a user can edit
Agent configuration and update it in their k8s cluster. It's quite generic,
independent of the telemetry type / deployment topology. There are two ways:
edit configmap embedded inside values.yaml of the Helm chart, or edit/update the
configmap directly. We'll describe both ways. There is also something to be said
about config reloader and checking for errors when invalid config is passed.