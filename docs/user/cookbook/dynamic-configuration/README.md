---
draft: true
---

# Dynamic Configuration Cookbook

The purpose of the cookbook is to guide you through common scenarios of using dynamic configuration. Each folder contains increasingly more complex use cases, but feel free to jump in wherever you feel appropriate.

## Basics

[Basics](01_Basics) covers
- [Structure](01_Basics/01_Structure.md) of how agent and server templates are loaded
- [Instances](01_Basics/02_Instances.md) of metrics and metrics instances are loaded
- [Integrations](01_Basics/03_Integrations.md) of how integrations are loaded
- [Logs and Traces](01_Basics/04_Logs_and_Traces.md) of how traces and logs are loaded

[Templates](02_Templates) covers
- [Looping](02_Templates/01_Looping.md) covers basic command usage and simple loops
- [Datasource](02_Templates/02_Datasources.md) covers usage of datasource which are external datastores you can use to pull in data
- [Datasources and Objects](02_Templates/03_Datasource_and_Objects.md) covers the usage of complex json objects

[Advanced Datasources](03_Advanced_Datasources) covers non file based datasources
- [AWS](03_Advanced_Datasources/01_AWS.md) covers querying EC2 for instances
