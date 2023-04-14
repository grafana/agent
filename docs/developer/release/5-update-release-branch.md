# Update Release Branch

The `release` branch is a special branch that is used for grafana cloud to point at our install scripts and example kubernetes manifests. This is not to be confused with `release-VERSION_PREFIX` created in [Create Release Branch](./1-create-release-branch.md)

## Before you begin

1. The release tag should exist from completing [Tag Release](./4-tag-release.md)

## Steps

1. Force push the release tag to the `release` branch

    ```
    git fetch
    git checkout main
    git branch -f release VERSION
    git push -f origin refs/heads/release
    ```

    > **NOTE**: This requires force push permissions on this branch. If this fails, reach out to one of the project maintainers for help. 
