---
title: "Designing in the Open"
weight: 1
---

# Designing in the Open

* Date: 2021-11-02
* Author: Robert Fratto (@rfratto)
* PR: [grafana/agent#1055](https://github.com/grafana/agent/pull/1055)

## Summary

Traditionally, design proposals for Grafana Agent tend to be done internally to
Grafana Labs by default. If the proposals are made public, it generally only
happens after internal consensus was already reached. This can cause a few
problems for external contributors:

* Community concerns are less likely to have an impact as core maintainers have
  already agreed on a solution.
* Initial design is gated to Grafana Labs employees.
* Historical design docs with context and discussions become hard to find both
  internally and even harder publicly.

In the spirit of fostering broader community participation, this document
proposes inverting the design process to best-effort being done in the open.

## Goals

* Outline a process to support public-first design documents
* Lower the barrier to entry to becoming a maintainer that is not employed at
  Grafana Labs
* Encourage a community to contribute input to design documents
* Encourage a community to contribute their own design documents

## Non-Goals

* Guarantee that everything will be made public. All contributors, regardless
  of company affiliation, are encouraged to design publicly when possible. We
  recognize that there will always be situations in which public design cannot
  happen, including (but not limited to) legal or security reasons.
* Identify strict rules for when an RFC or issue is appropriate.

## Terminology

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL
NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED",  "MAY", and
"OPTIONAL" in this document are to be interpreted as described in
[RFC 2119](https://datatracker.ietf.org/doc/html/rfc2119).

## Proposal

Public propsals come in two forms:

* RFCs (i.e., this document) for larger changes
* Issues proposals for smaller changes

This document does not prescribe when something is a "larger change" or a
"smaller change." We don't have enough experience yet to be able to
strongly define categories, but these guidelines can be used for a gut check:

* Larger changes require proposals that require thought and explore multiple
  options, OR
* Larger changes have a lasting effect on the project as a whole, OR
* Larger changes touch multiple systems throughout the project or dramatic
  changes to a single system.
* Smaller changes are the inverse of the above.

### RFCs

RFCs MUST be valid markdown composed of at least the following:

1. A title.
2. A bulletpoint list of metadata that MUST include original authorship date
   and the author's name. Author name SHOULD include their GitHub username as a
   parenthethical.
3. A summary section describing the context for what is being proposed.
4. A list of goals and non-goals for the proposal.
5. One or more sections dedicated to the proposal.

RFCs MUST be placed into the `docs/rfcs` folder of the Grafana Agent repository.

RFCs SHOULD have a `Terminology` section following the goals and non goals that
conform to [RFC 2119](https://datatracker.ietf.org/doc/html/rfc2119).

Sections dedicated to the proposal SHOULD include alternatives considered.

#### RFC IDs

The filename of the RFC MUST be prefixed with the ID of the RFC. If an RFC does
not yet have an ID, the ID MAY be omitted until one is assigned to it.

#### RFC Mutability

Once an RFC is accepted, all non-metadata information MUST remain immutable.
RFCs MAY be deprecated by further RFCs. If an RFC is deprecated, the older
RFC's metadata MUST be updated to point to what replaces it, and the newer RFC
MUST refer back to what it is replacing.

#### RFC Review

RFCs SHOULD be announced on the Grafana Agent's
[issue](https://github.com/grafana/agent/issues) page with a tag of `[RFC]`.
Review is performed by both maintiners and the public in form of inline code
review comments against the markdown and GitHub PR comments.

RFC PRs MUST NOT add or modify other files that are unrelated to the proposed
RFC.

### Issue Proposals

Issue proposals are the second form of proposal, useful for smaller proposals
that don't warrant an entire RFC.

Issue proposals are freeform, but SHOULD include:

* Background on what is being proposed
* The proposal itself
* If relevant, code or configuration examples which compliment the proposal

## Concerns

Grafana Loki previously tried to do public proposals similar to what is being
described here. It was found that reviewing RFCs via a Pull Request was clunky
and hard to manage.

This is an ongoing concern for this proposal, and while other projects manage it
successfully, it isn't clear if there are any suggested tools we could use to
make the review process easier for readers.

## Considered alternatives

A few existing public proposal processes have been examined for inspiration:

* [IETF's RFCs](https://www.ietf.org/standards/rfcs/)
* [Rust's RFCs](https://github.com/rust-lang/rfcs)
* [Joyent's Requests for Discussions](https://github.com/joyent/rfd)
* [OpenTelemetry's OTEPs](https://github.com/open-telemetry/oteps)
* [Kubernetes Enhancement Proposals (KEPs)](https://github.com/kubernetes/enhancements)

All of these processes are similar, but in the end, a mix of the IETF and Rust
workflows were used to derive this document.
