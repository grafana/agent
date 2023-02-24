# Releasing

This document describes the process of creating a release for the
`grafana/agent` repo. A release includes release assets for everything inside
the repository, including Grafana Agent and Grafana Agent Operator.

The processes described here are for v0.24.0 and above.

# Prerequisites

These [Prerequisites](./prerequisites.md) should be done by the release shepherd 
before taking any actions.

# Workflows

Once a release is scheduled, a release shepherd is determined. This person will be 
responsible for ownership of the following workflows:

- Release Candidate Publish
  - [Actions] 1-4
- Additional Release Candidate[s] Publish
  - [Actions] 2-4
- Stable Release Publish
  - [Actions] 2,TBD-TBD
- Patch Release Publish
  - [Actions] 2,TBD-TBD

# Actions

1. [Create Release Branch](./create-release-branch.md)
2. [Update Version in Code](./update-version-in-code.md)
3. [Tag Release](./tag-release.md)
4. [Publish Release Candidate](./publish-release-candidate.md)

[Actions]: #Actions

# Release Cycle

A typical release cycle is to have a Release Candidate published for at least 48 
hours followed by a Stable Release. 0 or more Patch Releases may occur between the Stable Release
and the creation of the next Release Candidate.

```mermaid
flowchart LR
    A(RCV) -->|>48 hours| B(SRV)
    B --> C(PRV 1)
    C --> D(PRV 2)
    D --> E(...)
    E --> F(New RCV)
```

RCV = Release Candidate Version

SRV = Stable Release Version

PRV = Patch Release Version