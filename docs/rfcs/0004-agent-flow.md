# Agent Flow - Agent Utilizing Components 

* Date: 2022-03-30
* Author: Matt Durham (@mattdurham)
* PR: [grafana/agent#1538](https://github.com/grafana/agent/pull/1538)
* Status: Draft

## Overarching Problem Statement

The Agents configuration and onboarding is difficult to use. Viewing the effect of configuration changes on telemetry data is difficult. Making the configuration simplier, composable and intuitive allieviates these concerns.


## Description

Agent Flow is intended to solve real world needs that the Grafana Agent team have idenfified in conversations with users and developers. 

These broadly include 

- Lack of introspection within the agent
    - Questions about what telemetry data are being sent
    - Are rules applying correctly?
    - How does filtering work?
- Users have been requesting additional capabilites and adding new features is hard due to coupling between systems, some examples include
    - Remote Write, Different Input Formats
    - Different output formats
    - Filtering ala Relabel Configs is complex and hard to figure out when they occur
- Lack of understanding how telemetry data moves through agent
    - Other systems use pipeline/extensions to allow users to understand how data moves through the system

# 1. Introduction and Goals 

This design document describes Agent Flow, which is a system for describing the execution and configuration of telemetry data, along with any secondary configuration required to configure the Grafana Agent. 

Agent Flow refers to both the execution, configuration and visual configurator of data flow.

The primary goals of Agent Flow are

1. Allow users to more easily understand the impact of their configuration
2. Allow users to collect metrics from all running integrations across a  set of agents without installing extra dependencies
3. Allow users to run components based on a dynamic environment
4. Allow developers to easily add components to the system
5. Ensure the Agent still maintains high performance on all platforms that the Agent currently supports
6. Flow configs must be easily machine-readable and machine-writable to support tooling, such as formatting and GUI

Goals of this design document

* Define high level concepts and goals
* Define high level concepts of the execution path

This document represents ideals and not technical implementation. 

Non Goals of this design document

* Define a technical implementation or configuration implementation

# 2. Broad Solution Path

Conversation around should the components be assembled via message passing, via expressions, or a hybrid approach.

**Note: Consider all examples pseudoconfig**

## 2.1 Expression Based

Expression based is writing expressions that allow referencing other components streams/outputs/values and using them directly. 

**Pros**

* Easier to Implement, evaluating expressions can map directly to existing config structs
* Components are more reusable, you can pass basic types arounds (string, int, bool)

**Cons**
* Harder for users to wire things together
* Harder to build a GUI for


## 2.2 Message Based

Message based is where components have no knowledge of other components and information is passed strictly via input and output streams. 

**Pros**

* Easier for users to understand the dependencies between components
* Easier to build a GUI for
* Ability to use more configuration formats easier

**Cons**

* More time consuming to implement, existing integrations/items would need to be componentized
* Larger type system needed
* More structured to keep the amount of types down

Messages would require a more rigid and well defined type structure. For instance for getting credentials from various sources and passing those credentials around we would want to avoid the following.

* MySQLCredentials
* RedisCredentials
* RemoteWriteCredentials
* MySQLCredentialsSourceS3
* MySQLCredentialsSourceVault
* MySQLCredentialsSourceConsul
* RedisCredentialsSourceS3
* RedisCredentialsSourceVault
* RedisCredentialsSourceConsul
* RemoteWriteCredentialsSourceS3
* RemoteWriteCredentialsSourceVault
* RemoteWriteCredentialsSourceConsul

And instead have one component that can read from many sources and outputs a single `Credential Type` and its up the destination component to intepret that correctly. 

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
                replacement = "root@$1"
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