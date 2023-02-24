# Update Version in Code

The codebase must be updated to reference the upcoming release tag whenever a new release is being made.

## Version

The version can be determined starting with the version prefix determined in [Create Release Branch](./create-release-branch.md) step 2.

- Release Candidate Version (RCV)

    The RCV will look like `[version prefix].x-rc.y`. For example, `v0.32.0-rc.0` is the first RCV for the v0.32.0 release.

- Stable Release Version (SRV)

    The SRV will look like `[version prefix].x`. For example, `v0.32.0` is the SRV for the v0.32.0 release.

- Patch Release Version (PRV)

    The PRV will look like `[version prefix].x`. For example, `v0.32.1` is the first PRV for the v0.32.0 release.

## Steps

1. Create a branch from `main`

    Note: This branch cannot be prefixed with `release-`

2. Update the `CHANGELOG.md`

    1. Modify Header:
        - First RCV or a PRV
            - Add a new header under `Main (unreleased)` for `[version] (YYYY-MM-DD)`
        - Additional RCV or SRV
            - Update the header `[previous RCV] (YYYY-MM-DD)` to `[version] (YYYY-MM-DD)`

    2. Move the unreleased changes included in the release branch from `Main (unreleased)` to `[version] (YYYY-MM-DD)`

    3. Update **appropriate** places in the codebase that have the previous version with the new version determined above.
    
        *This will require some tribal knowledge not documented here (yet).*

        NOTE: Please do not update the operations/helm directory. It is updated independently from Agent releases for now.

3. Update the version in code

    There are a number of places in code that the current version must be replaced with the new version.

4. Create a PR to merge to main (must be merged before continuing)

    See [here](https://github.com/grafana/agent/pull/2838/files) for an example PR for the first RCV.
    
    See [here](https://github.com/grafana/agent/pull/2873/files) for an example PR for a SRV.

5. Create a branch from `release-[version prefix]`
    
    Note: This branch cannot be prefixed with `release-`

6. Cherry pick the change commit from the merged PR in step 4 from main into the new branch from step 5.

7. Create a PR to merge to `release-[version prefix]` (must be merged before continuing)