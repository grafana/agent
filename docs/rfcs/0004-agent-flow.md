# This provided the basis for Agent Flow, and though not all the concepts/ideas will make it into flow, it is good to have the historical context for why we started down this path.

# Agent Flow - Agent Utilizing Components


* Date: 2022-03-30
* Author: Matt Durham (@mattdurham)
* PRs:
    * [grafana/agent#1538](https://github.com/grafana/agent/pull/1538) - Problem Statement
    * [grafana/agent#1546](https://github.com/grafana/agent/pull/1546) - Messages and Expressions

## Overarching Problem Statement

The Agents configuration and onboarding is difficult to use. Viewing the effect of configuration changes on telemetry data is difficult. Making the configuration simpler, composable and intuitive alleviates these concerns.


## Description

Agent Flow is intended to solve real world needs that the Grafana Agent team have identified in conversations with users and developers.

These broadly include:

- Lack of introspection within the agent
    - Questions about what telemetry data are being sent
    - Are rules applying correctly?
    - How does filtering work?
- Users have been requesting additional capabilities and adding new features is hard due to coupling between systems, some examples include
    - Remote Write, Different Input Formats
    - Different output formats
    - Filtering ala Relabel Configs is complex and hard to figure out when they occur
- Lack of understanding how telemetry data moves through agent
    - Other systems use pipeline/extensions to allow users to understand how data moves through the system

# 1. Introduction and Goals

This design document outlines Agent Flow, a system for describing a programmable pipeline for telemetry data.

Agent Flow refers to both the execution, configuration and visual configurator of data flow.

### Goals

* Allow users to more easily understand the impact of their configuration
* Allow users to collect integration metrics across a set of agents
* Allow users to run components based on a dynamic environment
* Allow developers to easily add components to the system
* Maintain high performance on all currently-supported platforms
* Produce machine-readable and machine-writable configs for tooling such as formatters or a GUI.

### Non-goals

* Discuss technical details: we instead focus on how a user would interact with a hypothetical implementation of Agent Flow.

# 2. Broad Solution Path

At a high level, Agent Flow:

* Breaks apart the existing hierarchical configuration file into reusable components
* Allows components to be connected, resulting in a programmable pipeline of telemetry data

This document considers three potential approaches to allow users to connect components together:

1. Message passing (i.e., an actor model)
2. Expressions (i.e., directly referencing the output of another component)
3. A hybrid of both messages and expressions

The Flow Should in general resemble a flowchart or node graph. The data flow diagram would conceptually look like the below, with each node being composable and connecting with other nodes.

```
┌─────────────────────────┐             ┌──────────────────┐          ┌─────────────────────┐       ┌───────────────────┐
│                         │      ┌─────▶│  Target Filter   │─────────▶│  Redis Integration  │──────▶│   Metric Filter   │──┐
│                         │      │      └──────────────────┘          └─────────────────────┘       └───────────────────┘  │
│    Service Discovery    │──────┤                                                                                         │
│                         │      │                                                                                         │
│                         │      │                                                                                         │
└─────────────────────────┘      │      ┌─────────────────┐           ┌──────────────────────┐                    ┌────────┘
                                 ├─────▶│  Target Filter  │──────────▶│  MySQL Integrations  │───────────┐        │
                                 │      └─────────────────┘           └──────────────────────┘           │        │
                                 │                                                                       │        │
                                 │       ┌─────────────────┐              ┌─────────────┐                │        │
                                 └──────▶│  Target Filter  │─────────────▶│   Scraper   │─────────────┐  │        │  ┌────────────────┐
                                         └─────────────────┘              └─────────────┘             └──┴┬───────┴─▶│  Remote Write  │
                                                                                                          │          └────────────────┘
                                                                                                          │
                                                                                                          │
┌──────────────────────────┐                                                                              │
│  Remote Write Receiver   │─────┐                                      ┌───────────────────────┐         │
└──────────────────────────┘     │                                ┌────▶│  Metric Transformer   │─────────┘
                                 │                                │     └───────────────────────┘
                                 │                                │
┌─────────────────────────┐      │      ┌────────────────────┐    │
│      HTTP Receiver      │──────┴─────▶│   Metric Filter    │────┘                           ┌──────────────────────────────────┐
└─────────────────────────┘             └────────────────────┘                                │    Global and Server Settings    │
                                                                                              └──────────────────────────────────┘
```

**Note: Consider all examples pseudoconfig**

## 2.1 Expression Based

Expression based is writing expressions that allow referencing other components streams/outputs/values and using them directly. Expressions allow referencing other fields, along with complex programming concepts. (functions, arithmetic ect). For instance `field1 = len(service_discover1.targets)`.

**Pros**

* Easier to Implement, evaluating expressions can map directly to existing config structs
* Components are more reusable, you can pass basic types around (string, int, bool) in addition to custom types

**Cons**
* Harder for users to wire things together
  * References to components are more complex, which may be harder to understand
* Harder to build a GUI for
  * Every field of a component is potentially dynamic, making it harder to represent visually


## 2.2 Message Based

Message based is where components have no knowledge of other components and information is passed strictly via input and output streams.

**Pros**

* Easier for users to understand the dependencies between components
* Easier to build a GUI for
    * Inputs and Outputs are well defined and less granular
    * Connections are made by connecting two components directly, compared to expressions which connect subsets of a component's output
* References between components are no more than strings, making the text-based representation language agnostic (e.g., it could be YAML, JSON, or any language)

**Cons**

* More time consuming to implement, existing integrations/items would need to be componentized
* Larger type system needed
* More structured to keep the amount of types down

Messages require a more rigid type structure to minimize the number of total components.

For example, it would be preferable to have a single `Credential` type that can be emitted by an s3, Vault, or Consul component. These components would then need to set a field that marks their output as a specific kind of Credential (such as Basic Auth or Bearer Auth).

If, instead, you had multiple Credential types, like `MySQLCredentials` and `RedisCredentials`, you would have the following components:

* Vault component for MySQL credentials
* Vault component for Redis credentials
* S3 component for MySQL credentials
* S3 component for Redis credentials
* (and so on)

## 2.3 Hybrid

## 2.4 Examples

### 2.4.1 Simple Example Mysql from Target Discovery

**Expression**

```
discovery "mysql_pods" {
    # some sort of config here to find pods
}


integration "mysql" {
  # Create one mysql integration for every element in the array here
  for_each = discovery.mysql_pods.targets

  # Each spawned mysql integration has its data_source_name derived from
  # the address label of the input target.
  data_source_name = "root@(${each.labels["__address__"]})"
}
```

**Message**

```
discovery "mysqlpods" {
    relabel_config {
        [
            {
                source = "__address__"
                match = "*mysql"
                action = "replace"
                replacement = "root@($1)"
            }
        ]
    }
}

# I think this would depend on convention, mysql would look at __address__ , and maybe optionally look for username/password
integration "mysql" {}

connections {
    [
        {
            source = mysqlpods
            destination = mysql
        }
    ]
}
```
