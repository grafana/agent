# Grafana Agent backwards compatibility 

* Date: 2023-05-25
* Author: Robert Fratto (@rfratto)
* PR: [grafana/agent#3981](https://github.com/grafana/agent/pull/3981)

Grafana Agent has been following [semantic versioning](https://semver.org/) since its inception.
After three years of development and 33 minor releases, the project is on trajectory to have a 1.0 release. 

In the context of semantic versioning, a 1.0 release indicates that future minor releases have backwards compatibility with older minor releases in the same major release; version 1.1 is backwards compatible with version 1.0. Having major and minor releases signals to users when an upgrade may take more time (major releases) and when they can upgrade with more confidence (minor releases).

However, Grafana Agent is a large project, with a large surface area for what may be considered part of the backwards compatibility guarantees. This proposal formally establishes what parts of Grafana Agent will be protected by backwards compatibility. 

## Goals 

- Set expectations for what is covered by backwards compatibility. 
- Set expectations for when upgrades to new major versions will be forced.

## Proposal 

Backwards compatibility means that a user can upgrade their version of Grafana Agent without needing to make any changes with the way they interact with Grafana Agent, provided that interaction is within scope of being covered by backwards compatibility.

### Scope of backwards compatibility   

The following will be protected by backwards compatibility between minor releases: 

- **User configuration**, including the syntax and functional semantics of the configuration file and command-line interface.

- **Versioned network APIs**, if any versioned APIs are introduced prior to the 1.0 release.

- **Telemetry data used in official dashboards**. This means that users will continue to be able to use the same set of dashboards we provide when upgrading minor releases.    

  - Official dashboards are dashboards in the repository's `operations/` [directory](../../operations/).

- **Externally importable Go packages**. If a user is importing our code as a dependency, they should be able to upgrade to a new minor release without having to make changes to their code.

  The backwards compatibility rules of these packages follow the same expectations as the [Go 1 compatibility][] expectations.

- **The scope of backwards compatibility**. Backwards compatibility is only defined for major version 1; we reserve the right to change the definition of backwards compatibility between major versions. 

If a breaking change is introduced in a minor change accidentally, and that breaking change is not covered by one of the [exceptions][] defined below, it is a bug. In these cases, a patch release should be introduced to undo the breaking change. 

[exceptions]: #exceptions-to-backwards-compatibility
[Go 1 compatibility]: https://go.dev/doc/go1compat

### Exceptions to backwards compatibility 

It's impossible to guarantee that full backwards compatibility is achieved. There are some exceptions which may cause a breaking change without a new major version:

- Non-stable functionality: Functionality which is explicitly marked as non-stable are exempt from backwards compatibility between minor releases.

  Non-stable functionality should be backwards compatible between patch releases, unless a breaking change is required for that patch release.

- Security: a breaking change may be made if a security fix requires making a breaking change. 

- Legal requirements: a breaking change may be made if functionality depends on software with a license incompatible with our own.

- Non-versioned network APIs: internal network APIs, such as the internal API used to drive the Flow web UI, are not subject to backwards compatibility guarantees.

- Undocumented behavior: relying on undocumented behavior may break between minor releases. 

- Upstream dependencies: part of the public API of Grafana Agent may directly expose the public API of an upstream dependency. In these cases, if an upstream dependency introduces a breaking change, we may be required to make a breaking change to our public API as well.   

- Other telemetry data: metrics, logs, and traces may change between releases. Only telemetry data which is used in official dashboards is protected under backwards compatibility.

### Avoiding major release burnout 

As a new major release implies a user must put extra effort into upgrading, it is possible to burn out users by releasing breaking changes too frequently. 

We will attempt to limit new major versions no more than once every 12 calendar months. This means that if Grafana Agent 1.0 was hypothetically released on August 4th, Grafana Agent 2.0 should not be released until at least August 4th of the following year. This is best-effort; if a new major release is required earlier, then we should not prevent ourselves from publishing such a release.

> **NOTE**: Here, "publishing a release" refers to creating a new versioned release associated with a Git tag and a GitHub release.
>
> Maintainers are free to queue breaking changes for the next major release in a branch at will.

Major releases should be aligned with breaking changes to the public API and not used as a way to hype a release. If hyping releases is required, there should be a version split between the API version of Grafana Agent and a project version (such as API v1.5, project version 2023.0).   

### Supporting previous major releases

When a new major release is published, the previous major release should continue to receive security and bug fixes for a set amount of time. Announcement of a new major release should be coupled with a minimum Long-Term Support (LTS) period for the previous major release. For example, we may choose to announce that Grafana Agent 0.X will continue to be supported for at least 12 months. 

LTS versions primarily receive security and bug fixes in the form of patch releases. New functionality in the form of minor releases is unlikely to be added to an LTS version, but may happen at the discretion of the project maintainers. 

Enabling LTS versions will give users additional time they may need to upgrade, especially if there is a significant amount of breaking changes to consider with the new major release. 

The support timeframe for an LTS version is not fixed, and may change between major releases. For example, version 0.X may receive at least 12 months of LTS, while version 1.X may receive at least 4 months of LTS. Project maintainers will need to decide how long to support previous major versions based on the difficulty for upgrading to the latest major version.
