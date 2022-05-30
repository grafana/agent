---
aliases:
- /docs/agent/latest/dynamic-configuration/dynamic/configuration
title: Dynamic Configuration
weight: 100
---

# Dynamic Configuration Cookbook

The purpose of this cookbook is to guide you through common scenarios of using dynamic configuration. Each section contains increasingly more complex use cases, but feel free to jump in wherever you feel appropriate.

As this

## Basics

The basic section covers
- The [Structure]({{< relref "./01_Basics/01_Structure.md" >}}) of how agent and server templates are loaded
- How [Instances]({{< relref "./01_Basics/02_Instances.md" >}}) of metrics and metrics instances are loaded
- How [Integrations]({{< relref "./01_Basics/03_Integrations.md" >}}) are loaded
- How [Logs and Traces]({{< relref "./01_Basics/04_Logs_and_Traces.md" >}}) of how traces and logs are loaded

## Templates
The Templates section includes

- [Looping]({{< relref "./02_Templates/01_Looping.md" >}}) with basic command usage and simple loops
- [Datasource]({{< relref "./02_Templates/02_Datasources.md" >}}) includes usage of external datastores you can use to pull in data as new data sources.
- [Datasources and Objects]({{< relref "./02_Templates/03_Datasources_and_Objects.md" >}}) covers the usage of complex json objects

## Advanced
The Advanced section includes
- The [AWS]({{< relref "./03_Advanced_Datasources/01_AWS.md" >}}) example that queries EC2 for instances.
