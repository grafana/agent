# Contributing

Grafana Agent uses GitHub to manage reviews of pull requests.

If you're planning to do a large amount of work, you should discuss your ideas
in an [issue][new-issue] or an [RFC][]. This will help you avoid unnecessary
work and surely give you and us a good deal of inspiration.

Pull requests can be opened immediately without an issue for trivial fixes or
improvements.

## Before Contributing

* Review the following code coding style guidelines:
  * [Go Code Review Comments][code-review-comments]
  * The _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments][best-practices]
  * The [Uber Go Style Guide][uber-style-guide]
* Sign our [CLA][], otherwise we're not able to accept contributions.

## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the
GitHub issue that you want to work on it. This is to prevent duplicated efforts
from contributors on the same issue.

Please check the [`good first issue`][good-first-issue] label to find issues
that are good for getting started. If you have questions about one of the
issues, with or without the tag, please comment on them and one of the
maintainers will clarify it. For a quicker response, contact us in the #agent
channel in our [community Slack][community-slack].

See next section for detailed instructions to compile the project. For quickly
compiling and testing your changes do:

```bash
# For building:
go build ./cmd/grafana-agent/
./grafana-agent -config.file=<config-file>

# For testing:
make lint test # Make sure all the tests pass before you commit and push :)
```

We use [`golangci-lint`](https://github.com/golangci/golangci-lint) for linting
the code.

As a last resort, if linting reports an issue and you think that the warning
needs to be disregarded or is a false-positive, you can add a special comment
`//nolint:linter1[,linter2,...]` before the offending line.

All our issues are regularly tagged with labels so that you can also filter
down the issues involving the components you want to work on.

## Compiling the Agent

To build Grafana Agent from source code, please install the following tools:

1. [Git](https://git-scm.com/)
2. [Go](https://golang.org/) (version 1.19 and up)
3. [Make](https://www.gnu.org/software/make/)
4. [Docker](https://www.docker.com/)

You can directly use the go tool to download and install the agent binary into your GOPATH:

    $ GO111MODULE=on go install github.com/grafana/agent/cmd/grafana-agent
    $ grafana-agent -config.file=your_config.yml

An example of the above configuration file can be found [here][example-config].

You can also clone the repository yourself and build using `make agent`:

    $ mkdir -p $GOPATH/src/github.com/grafana
    $ cd $GOPATH/src/github.com/grafana
    $ git clone https://github.com/grafana/agent.git
    $ cd agent
    $ make agent
    $ ./build/grafana-agent -config.file=your_config.yml

The Makefile provides several targets:

* `agent`: build the agent binary
* `test`: run the tests
* `lint`: run linting checks

### Compile on Linux
Compiling Grafana Agent on Linux requires extra dependencies:

* [systemd headers](https://packages.debian.org/sid/libsystemd-dev) for Promtail
   * Can be installed on Debian-based distributions with: ```sudo apt-get install libsystemd-dev```

## Pull Request Checklist

Changes should be branched off of the `main` branch. It's recommended to rebase
on top of `main` before submitting the pull request to fix any merge conflicts
that may have appeared during development.

PRs should not introduce regressions or introduce any critical bugs. If your PR
isn't covered by existing tests, some tests should be added to validate the new
code (note that 100% code coverage is _not_ a requirement). Smaller PRs are
more likely to be reviewed faster and easier to validate for correctness;
consider splitting up your work across multiple PRs if making a significant
contribution.

If your PR is not getting reviewed or you need a specific person to review it,
you can @-reply a reviewer asking for a review in the pull request or a
comment, or you can ask for a review on the Slack channel
[#agent](https://slack.grafana.com).

## Updating the changelog

We keep a [changelog](../../CHANGELOG.md) of code changes which result in new
or changed user-facing behavior.

Changes are grouped by change type, denoted by `### Category_Name`. The change
types are, in order:

1. Security fixes
2. Breaking changes
3. Deprecations
4. Features
5. Enhancements
6. Bugfixes
7. Other changes

Categories won't be listed if there's not any changes for that category.

When opening a PR which impacts user-facing behavior, contributors should:

1. Determine which changes need to be documented in the changelog (a PR may
   change more than one user-facing behavior).

2. If there are no other changes for that change type, add a header for it
   (e.g., `### Bugfixes`). Make sure to keep the order listed above.

3. Add relevant entries into the changelog.

When in doubt, look at a previous release for style and ordering examples.

### Changelog entry style tips

Change entries in the changelog should:

1. Be complete sentences, ending in a period. It is acceptable to use multiple
   complete sentences if one sentence can't accurately describe the change.
2. Describe the impact on the user which is reading the changelog.
3. Include credit to the GitHub user that opened the PR following the sentence.

For example:
`- Config file reading is now 1500% faster. (@torvalds)`

> Readers should be able to understand how a change impacts them. Default to
> being explicit over vague.
>
> * Vague: `- Fixed issue with metric names. (@ghost)`
> * Explicit: `- Fixed issue where instances of the letter s in metric names were replaced with z. (@ghost)`

## Dependency management

The Grafana Agent project uses [Go modules][go-modules] to manage dependencies
on external packages.

To add or update a new dependency, use the `go get` command:

```bash
# Pick the latest tagged release.
go get example.com/some/module/pkg@latest

# Pick a specific version.
go get example.com/some/module/pkg@vX.Y.Z
```

Tidy up the `go.mod` and `go.sum` files:

```bash
go mod tidy
```

You have to commit the changes to `go.mod` and `go.sum` before submitting the
pull request.

### Using forks

Using a fork to pull in custom changes must always be temporary.

PRs which add `replace` directives in go.mod to change a module to point to a
fork will only be accepted once an upstream PR is opened to officially move the
changes to the official module.

Contributors are expected to work with upstream to make their changes
acceptable, and remove the `replace` directive as soon as possible.

If upstream is unresponsive, consider choosing a different dependency or making
a hard fork (i.e., creating a new Go module with the same source).

[new-issue]: https://github.com/grafana/agent/issues/new
[RFC]: ../rfcs/0001-designing-in-the-open.md
[code-review-comments]: https://code.google.com/p/go-wiki/wiki/CodeReviewComments
[best-practices]: https://peter.bourgon.org/go-in-production/#formatting-and-style
[uber-style-guide]: https://github.com/uber-go/guide/blob/master/style.md
[CLA]: https://cla-assistant.io/grafana/agent
[good-first-issue]: https://github.com/grafana/agent/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22
[community-slack]: https://slack.grafana.com/
[example-config]: ../../cmd/grafana-agent/agent-local-config.yaml
[go-modules]: https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more
