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

The lifetime of a release branch is sheparded by a single person, picked from a
rota of project maintainers and contributors. The release shepard is
responsible for managing all releases that compose the release branch.

For a new release branch, the release shepard will:

1. Gather consensus on which commit should be used as a base for the release
   branch.

2. Create and push the release branch from the selected base commit.

3. Create a PR to cherry-pick additional commits into the release branch as
   needed.

4. Create a PR to [update code](#updating-code) for the upcoming release
   candidate. A new section in the changelog should be added for the release
   candidate, documenting all changes introduced by the release candidate.

4. Update the changelog with a new section for the upcoming release candidate,
   documenting changes that will be introduced in that release candidate.

5. Tag the release candidate using the tag naming convention `vX.Y.0-rc.N`.

6. Run the release candidate in a testing environment for at least 48 hours. If
   you do not have a testing environment, one can be spawned locally using the
   sample environments in `example/k3d`.

   During this period, no regressions or critical issues must be found.
   Discovered issues should be fixed via PRs to main. Return to step 3 after
   fixes are available, cherry-picking the fixes into the release branch and
   starting a new release candidate.

7. Create a PR to update code for the upcoming stable release. The changelog
   sections for the release candidates should be replaced with a single section
   for the stable release.

8. Create the stable release tag.

9. Force-push the `release` branch to point at the stable release tag. This
   branch is used to externally reference files in the repository for a stable
   release.

For patch releases, the release shepard will:

1. Create a PR to cherry-pick relevant bug fixes into the release branch.

2. Create a PR to update code for the upcoming patch release. A new changelog
   section should be dded for the patch release.

3. Create the patch release tag.

After the release shepard pushes a new tag, they must [publish the release](#publishing-the-release).

### Updating code

The codebase must be updated whenever a new release is being made to reference
the upcoming release tag:

* Modify `CHANGELOG.md` with a new version number and its release date.
* Go through the entire repository and find references to the previous release
  version, updating them to reference the new version.
* Run `make generate-manifests` and `make generate-dashboards` to update
  manifests in case they are stale.

NOTE: Any time CHANGELOG.md is updated for a release, it should first be done
via PR to the release branch, and then by a second PR to main.

### Merge freezes

Release shepards may request a merge freeze to main for any reason during the
release process.

### Publishing the Release

GitHub Actions will create release assets and open a release draft for every
pushed tag. To publish the release:

1. Go to the [GitHub releases page](https://github.com/grafana/agent/releases)
   and find the drafted release.

2. Edit the drafted release, copying and pasting *notable changes* from the
   CHANGELOG. Add a link to the CHANGELOG, noting that the full list of changes
   can be found there. Refer to other releases for help with formatting this.

3. Optionally, have other team members review the release draft if you wish
   to feel more comfortable with it.

4. Publish the release!

### Update Agent Operator Helm chart version

The [Grafana Agent Operator Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/agent-operator) 
needs to be manually updated after the release is published. These are the required steps:

1. Copy the content of the last CRDs into helm-charts: from agent's repo `production/operator/crds/` to
helm-charts' repo `charts/agent-operator/crds`.
2. Update references of agent-operator app version in helm-charts pointing to release version and bump helm chart version.
You can use this [pull-request](https://github.com/grafana/helm-charts/pull/1831) as reference.

## Maintaining older release branches

Older release branches are maintained on a best-effort basis. The release
shepard for that branch determines whether an older release branch have a new
patch release at their own discretion.
