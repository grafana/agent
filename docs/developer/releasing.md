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

For a **new release candidate**, the release shepherd will:

1. Gather consensus on which commit should be used as a base for the release
   branch.

2. Create and push the release branch from the selected base commit:

   * The name of the release branch should be the name of the stable version we intend 
   to release, such as `release-v0.31`.
   * Release branches do not contain the `-rc.N` release candidate suffix. This means
     there is no branch called `release-v0.31-rc.0`.
   * Release branches do not contain the patch version. This means there is no 
     branch called `release-v0.31.0` or `release-v0.31.4`.

3. Create a PR to cherry-pick additional commits into the release branch as
   needed.

4. Create a PR to [update code](#updating-code) on `main` using the `-rc.N` release
   candidate version.

5. Create a PR to cherry pick the update from step 4 to the release branch.

6. [Create a tag](#create-a-tag).

7. [Publish the release candidate on GitHub](#publishing-the-release)

8. Announce the release candidate on the community Slack channel.

9. Run the release candidate in a testing environment for at least 48 hours. If
   you do not have a testing environment, one can be spawned locally using the
   sample environments in `example/k3d`.

   During this period, no regressions or critical issues must be found.
   Discovered issues should be fixed via PRs to main. Return to step 3 after
   fixes are available, cherry-picking the fixes into the release branch and
   starting a new release candidate.

   The `rc` version of the next release candidate would have to be incremented - 
   e.g. from `v0.31.0-rc.0` to `v0.31.0-rc.1`.

If no new issues were found during the environment testing, the shepherd can continue with 
creating a **stable release**:

1. Create a PR to [update code](#updating-code) on `main` using the _stable_
   release version (i.e. without the `-rc.N` suffix).

2. Cherry pick the commit from step 1 to the release branch.

3. [Create a tag](#create-a-tag).

4. Force-push the `release` branch to point at the stable release tag. This
   branch is used to externally reference files in the repository for a stable
   release, e.g. the "[latest](https://grafana.com/docs/agent/latest/)" documentation.

5. Make sure the documentation is updated:
    1. A new version of the documentation should be visible on the 
    [version switcher](https://grafana.com/docs/versions/?project=/docs/agent/).
    2. The "[latest](https://grafana.com/docs/agent/latest/)" one should be identical to the latest version.
    3. Check the "Upgrade guide" to see if it includes the latest version.

6. [Publish the release on GitHub](#publishing-the-release)

7. [Update Agent Operator Helm chart version](#update-agent-operator-helm-chart-version)

8. Make sure [Homebrew is updated](#updating-homebrew).

9. Announce the release on community Slack channel.

For **patch releases**, the release shepherd will:

1. Merge the fixes we need for the patch to `main`.
   
2. Cherry-pick the fixes from `main` into the release branch.

3. Follow the same instructions as the ones for a stable release, starting from step 1.

Some notes on patch releases:

- When creating patch releases, there is no need for a release candidate.

- Changes made to patch releases are not listed in the CHANGELOG for the next stable version.

- The tag of the patch release does yet not exist at the time when we update code for the 
main branch to reference it.


### Updating code

The codebase must be updated to reference the upcoming release tag whenever a new release is being made.

NOTE: Any change done for a release should be done in `main` first and then moved to the release branch.

NOTE: Branches used for PRs should have a name that doesn't start with `release-`, 
otherwise branch protection rules apply to it. Alternatively, branches used for PRs 
can come from forks.

#### Update the CHANGELOG

When creating a release **candidate** (i.e. an `-rc.N` version):

* Add a new section in the CHANGELOG for the new release candidate and include 
the date of the release, for example, "v0.31.0-rc.0 (2023-01-26)".
* All items form the "unreleased" section will move to a new section for the upcoming release version.
* See [here](https://github.com/grafana/agent/pull/2838/files) for an example pull request.
* Sanity check that the CHANGELOG doesn't contain features which don't belong to it. Sometimes features get added to the wrong version in the CHANGELOG due to bad merges.

When creating a **stable** release:

* Replace the version from the release candidate section to be the stable release instead 
of the `-rc.N` version.
* We do not leave release candidates in the CHANGELOG once there is a stable version for them.
* Make sure to also update the **date** in the title.
* See [here](https://github.com/grafana/agent/pull/2873/files) for an example pull request.

When creating a **patch** release:

* Add a new section to the CHANGELOG for the patch release.
* In the CHANGELOG, **move** the fixes which were cherry picked onto the release branch 
from the "unreleased" section to the section for the new patch release.

#### Update the "Upgrade guide"
The "Upgrade guide" is located in `docs/sources/upgrade-guide/_index.md` and lists breaking changes 
and deprecated features relevant to each release. Make sure that it is updated similarly to the CHANGELOG.

#### Search and replace the old version with the new one

Go through the entire repository and find references to the previous release
  version, updating them to reference the new version **where necessary**.

NOTE: Please do not update the `operations/helm` directory. It is updated independently 
from Agent releases for now.

### Merge freezes

Release shepherds may request a merge freeze to main for any reason during the
release process.

### Create a tag
Remember to **checkout** and **pull** the release branch before creating the tag!

The tag naming conventions are [described here](#release-branches)

Example commands:
```
git checkout release-v0.31
git pull
git tag -s v0.31.2
git push origin v0.31.2
```

After the push double check that the tag on GitHub corresponds to the tip of the release branch on GitHub.

### Publishing the Release

GitHub Actions will create release assets and open a release draft for every
pushed tag.

> **WARNING**: We should never force push a tag after the publish button is pressed 
for a release.

To publish the release:

1. Go to the [GitHub releases page](https://github.com/grafana/agent/releases)
   and find the drafted release.

2. Edit the drafted release, copying and pasting *notable changes* from the
   CHANGELOG. Add a link to the CHANGELOG, noting that the full list of changes
   can be found there. Refer to other releases for help with formatting this.

3. At the bottom of the release page, tick the check box to "add a discussion" 
under the category for "announcements".

4. Also tick other boxes at the bottom of the release page:

      * For release candidates, tick the checkbox to "set as pre-release".

      * For stable releases and patch releases to the latest release branch, 
      tick the checkbox to "set as the latest release".

5. Optionally, have other team members review the release draft if you wish
   to feel more comfortable with it.

6. Publish the release!

> **NOTE**: Release candidates should be retained on the 
[GitHub releases page](https://github.com/grafana/agent/releases).
Please do not remove them.

### Update Agent Operator Helm chart version

The [Grafana Agent Operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator) 
needs to be manually updated after the release is published. These are the required steps:

1. Copy the content of the last CRDs into helm-charts: from agent's repo `production/operator/crds/` to
helm-charts' repo `charts/agent-operator/crds`.
2. Update references of agent-operator app version in helm-charts pointing to release version and bump 
helm chart version.
3. There is no need to update the README.md manually - running the 
[helm-docs](https://github.com/norwoodj/helm-docs) utility in the `charts/agent-operator` directory 
will update it automatically.

You can use this [pull-request](https://github.com/grafana/helm-charts/pull/1831) as reference.

### Updating Homebrew
For every Agent release we need to update the [homebrew/homebrew-core](https://github.com/Homebrew/homebrew-core) repository.

When a new [release](https://github.com/grafana/agent/releases) is published, GitHub actions 
are automatically triggered to create a PR for the above repository:
* If there are no issues, they will be merged automatically.
* The merge is done from the [grafanabot/homebrew-core](https://github.com/grafanabot/homebrew-core) fork.
* The GitHub actions will not fail if the PRs fail to merge.
* The release shepherd needs to manually open the GitHub actions, 
scroll to the end to find the PR and double check that it was merged.
* If the CI fails, the PR will not be able to merge. The release shepherd can then push a fix to 
the same branch as the one that the PR is open for.

## Maintaining older release branches

Older release branches are maintained on a best-effort basis. The release
shepherd for that branch determines whether an older release branch have a new
patch release at their own discretion.

Note that patching an old release branch works differently from patching the latest release:
- `main` would not be updated to reference the patch version
- The change applied to the old release branch would have to be listed in two sections 
in the CHANGELOG:
   1. The patch release's section
   2. The "unreleased" section
- When publishing the release on GitHub, it should **not** be marked as "the latest".
