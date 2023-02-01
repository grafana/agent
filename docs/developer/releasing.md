# Releasing

This document describes the process of creating a release for the
`grafana/agent` repo. A release includes release assets for everything inside
the repository, including Grafana Agent and Grafana Agent Operator.

The processes described here are for v0.24.0 and above.

## Release Branches

A release branch is created for every major or minor release. That release
branch is then used for all release candidates, the stable release, and all
patch releases.

For any given release branch, there will be the following tags:

* `vX.Y.0-rc.N`: release candidate N for vX.Y.0. At least one release candidate
  will be made for every release.
* `vX.Y.0`: the stable release for vX.Y.
* `vX.Y.Z`: patch release Z for vX.Y. A release branch may have zero or more
  patch releases.

Release branches follow a `release-v<major>.<minor>` naming convention (i.e.,
`release-v0.24`).

## Releaser Prerequisites

Each maintainer performing a release should perform the following steps once
before performing the release.

### Add Existing GPG Key to GitHub

First, Navigate to your user's
[SSH and GPG keys settings page](https://github.com/settings/keys). If the GPG
key for the email address used to commit is not present, do the following:

1. Run `gpg --armor --export <your email address>`
2. Copy the output.
3. In the settings page linked above, click "New GPG Key".
4. Copy and paste the PGP public key block.

### Signing Commits and Tags by Default

To avoid accidentally publishing a tag or commit without signing it, run the
following to ensure all commits and tags are signed:

```bash
git config --global commit.gpgSign true
git config --global tag.gpgSign true
```

#### macOS Signing Errors

If you are on macOS and using an encrypted GPG key, `gpg-agent` may be unable
to prompt you for your private key passphrase. This will be denoted by an error
when creating a commit or tag. To circumvent the error, add the following into
your `~/.bash_profile` or `~/.zshrc`, depending on which shell you are using:

```
export GPG_TTY=$(tty)
```

## Performing Releases

The lifetime of a release branch is shepherded by a single person, picked from a
rota of project maintainers and contributors. The release shepherd is
responsible for managing all releases that compose the release branch.

For a **new release branch**, the release shepherd will:

1. Gather consensus on which commit should be used as a base for the release
   branch.

2. Create and push the release branch from the selected base commit. 

   The name of the release branch should be the name of the stable version we intend 
   to release, e.g. `release-v0.31`. Release branches do not contain the `-rc.N` release 
   candidate suffix, i.e. there is no `release-v0.31-rc.0`.

3. Create a PR to cherry-pick additional commits into the release branch as
   needed.

4. Create a PR to [update code](#updating-code) using the `-rc.N` release
   candidate version.

5. Tag the release candidate using the tag naming convention `vX.Y.0-rc.N`.

6. Run the release candidate in a testing environment for at least 48 hours. If
   you do not have a testing environment, one can be spawned locally using the
   sample environments in `example/k3d`.

   During this period, no regressions or critical issues must be found.
   Discovered issues should be fixed via PRs to main. Return to step 3 after
   fixes are available, cherry-picking the fixes into the release branch and
   starting a new release candidate.

7. Create a PR to [update code](#updating-code) again - this time using the stable
   release version. 

8. Create the stable release tag.

9. Make sure [Homebrew is updated](#updating-homebrew).

10. Force-push the `release` branch to point at the stable release tag. This
   branch is used to externally reference files in the repository for a stable
   release, e.g. the "[latest](https://grafana.com/docs/agent/latest/)" documentation.

11. Make sure the documentation is updated:
    1. A new version of the documentation should be visible on the 
    [version switcher](https://grafana.com/docs/versions/?project=/docs/agent/).
    2. The "[latest](https://grafana.com/docs/agent/latest/)" one should be identical to the latest version.

12. [Publish the release](#publishing-the-release)

13. Post a comment on the community Slack channel to let the rest of the community 
know about the release. You can post about release candidates too.

For **patch releases**, the release shepherd will:

1. Create a PR to cherry-pick relevant bug fixes into the release branch.

2. Create a PR to update code for the upcoming patch release. A new changelog
   section should be dded for the patch release.

3. Create the patch release tag.

4. [Publish the release](#publishing-the-release)

### Updating code

The codebase must be updated whenever a new release is being made to reference the upcoming release tag.

#### Update the changelog

NOTE: Any time CHANGELOG.md is updated for a release, it should first be done
via PR to the release branch, and then by a second PR to main.

When creating a release **candidate** (an `-rc.N` version):

* Add a new section in the changelog for the new release candidate and include 
today's date. E.g. "v0.31.0-rc.0 (2023-01-26)".
* All items form the "unreleased" section will move to a new section for the upcoming release version.
* See [here](https://github.com/grafana/agent/pull/2838/files) for an example pull request.
* Sanity check that the changelog doesn't contain features which don't belong to it. Sometimes features get added to the wrong version in the changelog due to bad merges.

When creating a **stable** release:

* Replace the version from the release candidate section to be the stable release instead 
of the `-rc.N` version.
* We do not leave release candidates in the changelog once there is a stable version for them.
* Make sure to also update the **date** in the title.
* See [here](https://github.com/grafana/agent/pull/2873/files) for an example pull request.

#### Search and replace the old version with the new one

Go through the entire repository and find references to the previous release
  version, updating them to reference the new version **where necessary**.

NOTE: There are files such as 
"[pkg/operator/defaults.go](https://github.com/grafana/agent/blob/main/pkg/operator/defaults.go)", 
where the version sometimes should not be replaced but added to a list of versions.
At the time of this writing, `defaults.go` is the only such file.

#### Update defaults.go

Add the new version to the "[pkg/operator/defaults.go](https://github.com/grafana/agent/blob/main/pkg/operator/defaults.go)" file. If there is a release candidate (`-rc.N`) version, remove it.

#### Update manifests and dashboards

Run `make generate-manifests` and `make generate-dashboards` to update
manifests in case they are stale.

### Merge freezes

Release shepherds may request a merge freeze to main for any reason during the
release process.

### Publishing the Release

GitHub Actions will create release assets and open a release draft for every
pushed tag. To publish the release:

1. Go to the [GitHub releases page](https://github.com/grafana/agent/releases)
   and find the drafted release.

2. Edit the drafted release, copying and pasting *notable changes* from the
   CHANGELOG. Add a link to the CHANGELOG, noting that the full list of changes
   can be found there. Refer to other releases for help with formatting this.

3. Tick the appropriate boxes at the bottom of the release page:

      * For release candidates:

         1. Tick the checkbox to "set as pre-release".

      * For stable releases:

         1. Tick the checkbox to "set as the latest release".

         2. Tick the check box to "add a discussion" under the category for "announcements".

4. Optionally, have other team members review the release draft if you wish
   to feel more comfortable with it.

5. Publish the release!

NOTE: Release candidates should be retained on the 
[GitHub releases page](https://github.com/grafana/agent/releases).
Please do not remove them.

### Update Agent Operator Helm chart version

The [Grafana Agent Operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator) 
needs to be manually updated after the release is published. These are the required steps:

1. Copy the content of the last CRDs into helm-charts: from agent's repo `production/operator/crds/` to
helm-charts' repo `charts/agent-operator/crds`.
2. Update references of agent-operator app version in helm-charts pointing to release version and bump helm chart version.
You can use this [pull-request](https://github.com/grafana/helm-charts/pull/1831) as reference.

### Updating Homebrew
[Homebrew](https://brew.sh/) is a package manager for MacOS. It installs binaries 
(aka "[bottles](https://docs.brew.sh/Bottles)") from either the main Homebrew repository 
or a third-party one (aka a "[tap](https://docs.brew.sh/Taps)").

For every Agent release we need to update two Brew repositories:
* [homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core) is the main one which 
Brew installations use by default
* [grafana/homebrew-grafana](https://github.com/grafana/homebrew-grafana) is a "tap" - 
a Third-Party Repository from Grafana.

When a release tag is published to the Agent repository, GitHub actions are triggered automatically 
to update both of the above repositories. PRs are created automatically:
* If there are no issues, they will be merged automatically.
* The merges are done from these two grafanabot forks:
    * [grafanabot/homebrew-core](https://github.com/grafanabot/homebrew-core)
    * [grafanabot/homebrew-grafana](https://github.com/grafanabot/homebrew-grafana/tree/master)
* The GitHub actions will not fail if the PRs fail to merge.
* The release shepherd needs to manually open the GitHub actions, 
scroll to the end to find the PR and double check that it was merged.
* If the CI fails the PR is not able to merge, the release shepherd can push a fix to 
the same branch as the one that the PR is open for.

## Maintaining older release branches

Older release branches are maintained on a best-effort basis. The release
shepherd for that branch determines whether an older release branch have a new
patch release at their own discretion.
