# Update Release Branch

The `release` branch is a special branch that is used for grafana cloud to point at our install scripts and example kubernetes manifests. This is not to be confused with `release-[version prefix]` created in [Create Release Branch](./1-create-release-branch.md)

## Steps

1. Force push the release tag to the `release` branch

    ```
    git checkout main
    git branch -f release [release version]
    git push -f origin refs/heads/release
    ```

    *NOTE: This requires force push permissions on this branch so you may need an assist doing this step*