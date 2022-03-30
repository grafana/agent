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

1. Allow users to easily configure the system
2. Allow users to understand the system
3. Allow users to inspect the system
4. Allow developers to easily add components to the system
4. Ensure the Agent still maintains high performance on all platforms that the Agent currently supports

Goals of this design document

* Define high level concepts and goals
* Define high level concepts of the execution path

This document represents ideals and not technical implementation. 

Non Goals of this design document

* Define a technical implementation or configuration implementation
