# Status and future plans for the Grafana Agent Operator

* Date: 2022-08-17
* Author: Craig Peterson (@captncraig)
* PR: [grafana/agent#2046](https://github.com/grafana/agent/pull/2046)

## Summary

A recent [draft rfc](https://github.com/grafana/agent/pull/1565) discussed the possibility of deprecating the Agent Operator. Although we ultimately closed that proposal, there are still valid concerns in the community about the long-term support to be expected around the operator.
In the interest of full transparency, we'd like to lay out our goals and vision for the project and reaffirm our commitment to it.

## Goals of the operator

The operator serves two primary functions:

1. Allow users to reuse the same monitoring CRDs provided by the [Prometheus Operator](), such as `PodMonitor` and `ServiceMonitor`. This is important to allow dynamic monitoring of kubernetes components, especially in many environments where monitoring configuration is divided among multiple teams.
2. Allow the Agent itself to be installed and configured using `GrafanaAgent` and `MetricsInstance` CRDs. This often simplifies deployments, and allows a declarative configuration style.

These two goals are somewhat independent of one another. Both of these use cases are important to us, and we are committed to supporting them into the future.

## Difficulties of the operator

The operator is a fairly complex piece of code, and has been slower than some other packages to receive updates and new Agent features. It has been received less documentation attention than other areas, and thus is often misunderstood. These are all areas we hope to improve over the next few months.

## Future Plans for the operator

1. We intend to bring full support for the Prometheus Operator CRDs into the Grafana Agent itself in the coming months. That will make a good deal of the core functionality of the Operator available to all Agent deployments, whether created by the operator or a helm chart, or whatever other method is chosen. That should also bring some performance and stability improvements, such as fewer agent reloads.
2. Doing that will allow us to reduce the scope of the operator itself while fully maintaining backward compatibility.
3. The operator will then be primarily useful for creating and configuring Grafana Agent instances declaratively. We can then potentially look into alternatives for generating Agent deployment manifests (helm charts, jsonnet libraries, etc..) as our primary recommendation, but will remain mindful that we are committed to the Operator, and will make reasonable efforts to maintain backward compatibility as much as possible.

## Beta status

The Grafana Agent Operator is still considered beta software. It has received a better reception than anticipated, and is now an important part of the Agent project. We are committed to supporting the Operator into the future, but are going to leave the beta designation in place while making larger refactorings as described above. We make efforts to avoid breaking changes, and hope that custom resource definitions will remain compatible, but it is possible some changes will be necessary. We will make every effort to justify and communicate such scenarios as they arise.

Once we are confident we have an Operator we are happy with and that the resource definitions are stable, we will revisit the beta status as soon as we can.
