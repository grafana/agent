# Tag Release

A tag is required to create GitHub artifacts and as a prerequisite for publishing.

## Prerequisites

All required commits for the release should exist on the release branch. This includes functionality and documentation such as the `CHANGELOG.md`. All versions in code should have already been updated.

## Steps

1. Make sure you are up to date on the release branch (git checkout, fetch and pull).

2. Tag the release.

    The release version was previously determined in [Update Version in Code](./3-update-version-in-code.md).

    Example commands:

    ```
    git tag -s [release version]
    git push origin [release version]
    ```

3. After a tag has been pushed, GitHub Actions will create release assets and open a release draft for every pushed tag.

    - This will take ~20-40 minutes.
    - You can monitor this by viewing the drone build on the commit for the release tag.

    *NOTE: homebrew may fail, this is OK*