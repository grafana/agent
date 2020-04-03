# Maintainer's Guide

This document provides relevant instructions for maintainers of the Grafana
Cloud Agent.

## Releasing

### Prerequisites

Each maintainer performing a release should perform the following steps once
before releasing the Grafana Cloud Agent.

#### Add Existing GPG Key to GitHub

First, Navigate to your user's [SSH and GPG keys settings
page](https://github.com/settings/keys). If the GPG key for the email address
used to commit with Grafana Cloud Agent is not present, follow these
instructions to add it:

1. Run `gpg --armor --export <your email address>`
2. Copy the output.
3. In the settings page linked above, click "New GPG Key".
4. Copy and paste the PGP public key block.

#### Signing Commits and Tags by Default

To avoid accidentally publishing a tag or commit without signing it, you can run
the following to ensure all commits and tags are signed:

```bash
git config --global commit.gpgSign true
git config --global tag.gpgSign true
```

##### macOS Signing Errors

If you are on macOS and using an encrypted GPG key, the `gpg-agent` may be
unable to prompt you for your private key passphrase. This will be denoted by an
error when creating a commit or tag. To circumvent the error, add the following
into your `~/.bash_profile` or `~/.zshrc`, depending on which shell you are
using:

```
export GPG_TTY=$(tty)
```

## Performing the Release

1. Create a new branch to update `CHANGELOG.md` and references to version
   numbers across the entire repository (e.g. README.md in the project root).
2. Modify `CHANGELOG.md` with the new version number and its release date.
3. List all the merged PRs since the previous release. This command is helpful
   for generating the list (modifying the date to the date of the previous release): `curl https://api.github.com/search/issues?q=repo:grafana/agent+is:pr+"merged:>=2019-08-02" | jq -r ' .items[] | "* [" + (.number|tostring) + "](" + .html_url + ") **" + .user.login + "**: " + .title'`
4. Go through the entire repository and find references to the previous release
   version, updating them to reference the new version.
5. *Without creating a tag*, create a commit based on your changes and open a PR
   for updating the release notes.
6. Merge the changelog PR.
7. Create a new tag for the release.
    1. Once this step is done, the CI will be triggered to create release
       artifacts and publish them to a draft release. The tag will be made
       publicly available immediately.
    2. Run the following to create the tag:

       ```bash
       RELEASE=v1.2.3 # UPDATE ME to reference new release
       git checkout master # If not already on master
       git pull
       git tag -s $RELEASE -m "release $RELEASE"
       git push origin $RELEASE
       ```
8. Watch GitHub Actions and wait for all the jobs to finish running.

## Publishing the Release Draft

Once the steps are completed, you can publish your draft!

1. Go to the [GitHub releases page](https://github.com/grafana/agent/releases)
   and find the drafted release.
2. Edit the drafted release, copying and pasting *notable changes* from the
   CHANGELOG. Add a link to the CHANGELOG, noting that the full list of changes
   can be found there. Refer to other releases for help with formatting this.
3. Optionally, have other team members review the release draft so you feel
   comfortable with it.
4. Publish the release!

## Updating Release Branch

The `release` branch should always point at the SHA of the commit of the latest
release tag. This is used so that the install instructions can be generic and
made to always install the latest released version.

Update the release branch by fast-forwarding it to the appropriate SHA (matching
the latest tag) and pushing it back upstream.

## `grafana/prometheus` Maintainence

Grafana Labs is using the Agent for their internal monitoring and want to take
advantage of the Agent to proof-of-concept additions to Prometheus before they
get moved upstream. A `grafana/prometheus` repository is maintained by Grafana
Labs where all non-trivial changes will go first. Doing this allows for getting
Cortex-specific changes moving along faster and providing strong evidence
towards its usefulness and correctness when it becomes part of an upstream PR.

We are commiting ourselves to doing the following:

1. Always use a recent Prometheus release: the Agent will always vendor a
   recent Prometheus release and not Prometheus master. We want the Agent's
   Prometeheus roots to be stable.
2. Keep changes mergeable upstream: we want to continue to be good OSS citizens,
   and we intend that all features we add to our Prometheus repository will
   become an upstream PR. We will maintain our repository in a way that supports
   doing this.
3. Reduce code drift: The code the Agent uses on top of Prometheus will be
   layered on top of a Prometheus release rather than sandwiched in between.

Maintainence of the `grafana/prometheus` repository revolves around feature
branches (named `feat-SOME-FEATURE`) and release branches (named
`release-vX.Y.Z-grafana`). The release branches will always use the same release
version as the `prometheus/prometheus` release it is based off of.

By adding features to the `grafana/prometheus` repository first, we are
committing ourselves to extra maintenance of features that have not yet been
merged upstream. Feature authors will have to babysit their features to
coordinate with the Prometheus release schedule to always be compatible. One the
feature is merged upstream, this burden of maintaining is made easier as the
Prometheus team can more easily sync breaking changes across the whole
repository.

We are purposefully carrying this extra burden because we intend to ultimately
make Prometheus better and contribute all of our enhancements upstream. We want
to strive to benefit the Prometheus ecosystem at large.

### Creating a New Feature

For `grafana/prometheus` maintainers to create a new feature, they will do the
following:

1. Create a feature branch in `grafana/prometheus` based on the latest release
   tag that `grafana/prometheus` currently has a release branch for. The feature
   branch should follow the naming convention `feat-<feature name>`.
2. Implement the feature and open a PR to merge the feature branch into the
   associated `grafana/prometheus` release branch.
3. Once the release branch is updated, open a PR to update `grafana/agent` to
   use the latest release branch SHA.

### Updating an Existing Feature

If a feature branch that was already merged to a release branch needs to be
updated for any reason:

1. Push directly to the feature branch or open a PR to merge changes into that
   feature branch.
2. Open a PR to merge the new changes from the feature branch into the
   associated release branch.
3. Once the release branch is updated, open a PR to update `grafana/agent` by vendoring
   the changes using the latest release branch SHA.

### Handling New Upstream Release

When a new upstream `prometheus/prometheus` release is available, we must go
through the following process:

1. Create a new `grafana/prometheus` release branch named
   `release-X.Y.Z-grafana`.
2. For all feature branches still not merged upstream, rebase them on top of the
   latest release. Force push them to update the `grafana/prometheus` release
   branch.
3. Create one or more PRs to introduce the features into the newly created
   release branch.

Once this process is completed, the previous release branch in
`grafana/prometheus` is considered stale and will no longer be updated.

### Updating the Agent's vendor

The easiest way to do this is the following:

1. Edit `go.mod` and change the replace directive to the release branch name.
2. Run `go mod vendor && go mod tidy`.
3. Commit and open a PR.

### Gotchas

If the `grafana/prometheus` feature is incompatible with the upstream master
branch, an upstream PR cannot be merged due to the merge conflicts. There are a
few ways this can be handled at the feature author's discrection:

1. Rebase the feature branch to `prometheus/prometheus` master so it can be
   merged upstream. Doing this means that we cannot have the feature in th agent
   until a new upstream release is available containing the feature.
2. Wait until a new `prometheus/prometheus` release is available and rebase the
   feature branch on top. The upstream PR will now be compatible with master,
   but this window is small and may change at any time. This makes it slower to
   get the feature merged upstream.
3. Create a new feature branch based off of master and open a PR for that
   feature branch. This adds extra maintenance burden on the feature author as
   they now have to mirror changes across two feature branches.

### Open Questions

If two feature branches depend on one another, there are two suggested solutions
to handling this:

- Have a combined feature branch (like an "epic" branch) that contains a set of
  multiple related features. All features within that "epic" branch are merged
  directly to the combined branch rather than individual feature branches.
- Keep an ordered list of features that should be merged to a release branch in
  the order of dependency.
