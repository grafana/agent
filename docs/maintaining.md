# Maintainer's Guide

This document provides relevant instructions for maintainers of the Grafana
Agent.

## Master Branch Rename

The `master` branch was renamed to `main` on 17 Feb 2021. If you have already
checked out the repository, you will need to update your local environment:

```
git branch -m master main
git fetch origin
git branch -u origin/main main
```

## Releasing

### Prerequisites

Each maintainer performing a release should perform the following steps once
before releasing the Grafana Agent.

#### Prerelease testing

For testing a release, run the [K3d example](../example/k3d/README.md) locally.
Let it run for about 90 minutes, keeping an occasional eye on the Agent
Operational dashboard (noting that metrics from the scraping service will take
time to show up). After 90 minutes, if nothing has crashed and you see metrics
for both the scraping service and the non-scraping service, the Agent is ready
for release.

#### Add Existing GPG Key to GitHub

First, Navigate to your user's [SSH and GPG keys settings
page](https://github.com/settings/keys). If the GPG key for the email address
used to commit with Grafana Agent is not present, follow these
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

If you are performing a release for a release candidate, skip straight to step
8.

1. Create a new branch to update `CHANGELOG.md` and references to version
   numbers across the entire repository (e.g. README.md in the project root).
2. Modify `CHANGELOG.md` with the new version number and its release date.
3. Add a new section in `CHANGELOG.md` for `Main (unreleased)`.
4. Go through the entire repository and find references to the previous release
   version, updating them to reference the new version.
5. Run `make example-kubernetes` and `make example-dashboards` to update
   manifests in case they are stale.
6. *Without creating a tag*, create a commit based on your changes and open a PR
   for updating the release notes.
7. Merge the changelog PR.
8. Create a new tag for the release.
    1. After following step 2, the CI will be triggered to create release
       artifacts and publish them to a draft release. The tag will be made
       publicly available immediately.
    2. Run the following to create the tag:

       ```bash
       RELEASE=v1.2.3 # UPDATE ME to reference new release
       git checkout main # If not already on main
       git pull
       git tag -s $RELEASE -m "release $RELEASE"
       git push origin $RELEASE
       ```
9. Watch GitHub Actions and wait for all the jobs to finish running.

## Publishing the Release Draft

After this final set of steps, you can publish your draft!

1. Go to the [GitHub releases page](https://github.com/grafana/agent/releases)
   and find the drafted release.
2. Edit the drafted release, copying and pasting *notable changes* from the
   CHANGELOG. Add a link to the CHANGELOG, noting that the full list of changes
   can be found there. Refer to other releases for help with formatting this.
3. Optionally, have other team members review the release draft if you wish
   to feel more comfortable with it.
4. Publish the release!

The release isn't done yet! Keep reading for the final step.

## Updating Release Branch

If the release you are performing is a _stable release_ (i.e., not a release
candidate), the `release` branch must be updated to the SHA of the latest stable
release tag. This is used so that installation instructions can be generic and
made to always install the latest released version. Otherwise, if the release is
non-stable, the `release` branch should be left unmodified.

Update the release branch by fast-forwarding it to the appropriate SHA (matching
the latest tag) and pushing it back upstream.

## `grafana/prometheus` Maintenance

Grafana Labs includes the Agent as part of their internal monitoring, running it
alongside Prometheus. This gives an opportunity to utilize the Agent to
proof-of-concept additions to Prometheus before they get moved upstream. A
`grafana/prometheus` repository maintained by Grafana Labs holds non-trivial
and experimental changes. Having this repository allows for experimental features to
be vendored into the Agent and enables faster development iteration. Ideally,
this experimental testing can help serve as evidence towards usefulness and
correctness when the feature becomes proposed upstream.

We are committing ourselves to doing the following:

1. Keep changes mergeable upstream: we want to continue to be good OSS citizens,
   and we intend that all features we add to our Prometheus repository will
   become an upstream PR. We will maintain our repository in a way that supports
   doing this.
2. Always vendor a branch from `grafana/prometheus` based off of a recent Prometheus
   stable release; we want the Agent's Prometheus roots to be stable.
3. Reduce code drift: The code the Agent uses on top of Prometheus will be
   layered on top of a Prometheus release rather than sandwiched in between.
4. Keep the number of experimental changes not merged upstream to a minimum. We're
   not trying to fork Prometheus.

Maintenance of the `grafana/prometheus` repository revolves around feature
branches (named `feat-SOME-FEATURE`) and release branches (named
`release-vX.Y.Z-grafana`). The release branches will always use the same release
version as the `prometheus/prometheus` release it is based off of.

By adding features to the `grafana/prometheus` repository first, we are
committing ourselves to extra maintenance of features that have not yet been
merged upstream. Feature authors will have to babysit their features to
coordinate with the Prometheus release schedule to always be compatible. Maintenance
burden becomes lightened once each feature is upstreamed as breaking changes will
no longer happen out of sync with upstream changes for the respective upstreamed
feature.

We are purposefully carrying this extra burden because we intend to ultimately
make Prometheus better and contribute all of our enhancements upstream. We want
to strive to benefit the Prometheus ecosystem at large.

### Creating a New Feature

Grafana Labs developers should try to get all features upstreamed *first*. If
it's clear the feature is experimental or more unproven than the upstream team
is comfortable with, developers should then create a downstream
`grafana/prometheus` feature branch.

For `grafana/prometheus` maintainers to create a new feature, they will do the
following:

1. Create a feature branch in `grafana/prometheus` based on the latest release
   tag that `grafana/prometheus` currently has a release branch for. The feature
   branch should follow the naming convention `feat-<feature name>`.
2. Implement the feature and open a PR to merge the feature branch into the
   associated `grafana/prometheus` release branch.
3. After updating the release branch, open a PR to update `grafana/agent` to
   use the latest release branch SHA.

### Updating an Existing Feature

If a feature branch that was already merged to a release branch needs to be
updated for any reason:

1. Push directly to the feature branch or open a PR to merge changes into that
   feature branch.
2. Open a PR to merge the new changes from the feature branch into the
   associated release branch.
3. After updating the release branch, open a PR to update `grafana/agent` by
   vendoring the changes using the latest release branch SHA.

### Handling New Upstream Release

When a new upstream `prometheus/prometheus` release is available, we must go
through the following process:

1. Create a new `grafana/prometheus` release branch named
   `release-X.Y.Z-grafana`.
2. For all feature branches still not merged upstream, rebase them on top of the
   newly created branch. Force push them to update the `grafana/prometheus`
   feature branch.
3. Create one or more PRs to introduce the features into the newly created
   release branch.

Once a new release branch has been created, the previous release branch in
`grafana/prometheus` is considered stale and will no longer receive updates.

### Updating the Agent's vendor

The easiest way to do this is the following:

1. Edit `go.mod` and change the replace directive to the release branch name.
2. Update `README.md` in the Agent to change which version of Prometheus
   the Agent is vendoring.
2. Run `go mod tidy && go mod vendor`.
3. Commit and open a PR.

### Gotchas

If the `grafana/prometheus` feature is incompatible with the upstream
`prometheus/prometheus` master branch, merge conflicts would prevent
an upstream PR from being merged. There are a few ways this can be handled
at the feature author's discretion:

When this happens, downstream feature branch maintainers should wait until
a new `prometheus/prometheus` release is available and rebase their feature
branch on top of the latest release. This will make the upstream PR compatible
with the master branch, though the window of compatibility is unpredictable
and may change at any time.

If it proves unfeasible to get a feature branch merged upstream within the
"window of upstream compatibility," feature branch maintainers should create
a fork of their branch that is based off of master and use that master-compatible
branch for the upstream PR. Note that this means any changes made to the feature
branch will now have to be mirrored to the master-compatible branch.

### Open Questions

If two feature branches depend on one another, a combined feature branch
(like an "epic" branch) should be created where development of interrelated
features go. All features within this category go directly to the combined
"epic" branch rather than individual branches.
