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
