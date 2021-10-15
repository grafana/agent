# Contributing

Grafana Agent uses GitHub to manage reviews of pull requests.

* If you are a new contributor see: [Steps to Contribute](#steps-to-contribute)

* If you have a trivial fix or improvement, go ahead and create a pull request.

* If you plan to do something more involved, first discuss your ideas
  in an [issue](https://github.com/grafana/agent/issues/new). This will avoid unnecessary work and surely give you and us a good deal
  of inspiration.

* Relevant coding style guidelines are the [Go Code Review
  Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments)
  and the _Formatting and style_ section of Peter Bourgon's [Go: Best
  Practices for Production
  Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).

* Be sure to sign our [CLA](https://cla-assistant.io/grafana/agent).


## Steps to Contribute

Should you wish to work on an issue, please claim it first by commenting on the GitHub issue that you want to work on it. This is to prevent duplicated efforts from contributors on the same issue.

Please check the [`good first issue`](https://github.com/grafana/agent/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) label to find issues that are good for getting started. If you have questions about one of the issues, with or without the tag, please comment on them and one of the maintainers will clarify it. For a quicker response, contact us in the #agent channel in our [community Slack](https://slack.grafana.com/).

See next section for detailed instructions to compile the project. For quickly compiling and testing your changes do:
```
# For building.
go build ./cmd/agent/
./agent -config.file=<config-file>

# For testing.
make test         # Make sure all the tests pass before you commit and push :)
```

We use [`golangci-lint`](https://github.com/golangci/golangci-lint) for linting the code. If it reports an issue and you think that the warning needs to be disregarded or is a false-positive, you can add a special comment `//nolint:linter1[,linter2,...]` before the offending line. Use this sparingly though, fixing the code to comply with the linter's recommendation is in general the preferred course of action.

All our issues are regularly tagged so that you can also filter down the issues involving the components you want to work on.

## Compiling the Agent

To build Grafana Agent from source code, please install the following tools:

1. [git](https://git-scm.com/)
2. [go](https://golang.org/) (version 1.1X and up)
3. [make](https://www.gnu.org/software/make/)
4. [docker](https://www.docker.com/)

You can directly use the go tool to download and install the agent binary into your GOPATH:

    $ GO111MODULE=on go install github.com/grafana/agent/cmd/agent
    $ agent -config.file=your_config.yml

An example of the above configuration file can be found [here](TODO).

You can also clone the repository yourself and build using `make agent`:

    $ mkdir -p $GOPATH/src/github.com/grafana
    $ cd $GOPATH/src/github.com/grafana
    $ git clone https://github.com/grafana/agent.git
    $ cd agent
    $ make agent
    $ ./agent -config.file=your_config.yml

The Makefile provides several targets:

* `agent`: build the agent binary
* `test`: run the tests
* `lint`: will run golangci-lint


## Pull Request Checklist

* Branch from the main branch and, if needed, rebase to the current main branch before submitting your pull request. If it doesn't merge cleanly with main you may be asked to rebase your changes.

* Commits should be as small as possible, while ensuring that each commit is correct independently (i.e., each commit should compile and pass tests).

* If your patch is not getting reviewed or you need a specific person to review it, you can @-reply a reviewer asking for a review in the pull request or a comment, or you can ask for a review on the Slack channel [#agent](https://slack.grafana.com).

* Add tests relevant to the fixed bug or new feature.

## Dependency management

The Grafana Agent project uses [Go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) to manage dependencies on external packages.

To add or update a new dependency, use the `go get` command:

```bash
# Pick the latest tagged release.
go install example.com/some/module/pkg@latest

# Pick a specific version.
go install example.com/some/module/pkg@vX.Y.Z
```

Tidy up the `go.mod` and `go.sum` files:

```bash
# The GO111MODULE variable can be omitted when the code isn't located in GOPATH.
GO111MODULE=on go mod tidy
```

You have to commit the changes to `go.mod` and `go.sum` before submitting the pull request.