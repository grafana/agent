# Update Version in Code

This Action will typically follow [Create Release Branch](./create-release-branch.md) or if it is time to start a Stable Release or Patch Release.

The codebase must be updated to reference the upcoming release tag whenever a new release is being made.

## Release Version

The release version prefix was previously determined in [Create Release Branch](./create-release-branch.md).

- Release Candidate Version (RCV)

    - The RCV will look like `[release version prefix].x-rc.y`.
    - For example, `v0.32.0-rc.0` is the first RCV for the v0.32.0 release.

- Stable Release Version (SRV)

    - The SRV will look like `[release version prefix].0`.
    - For example, `v0.32.0` is the SRV for the v0.32.0 release.

- Patch Release Version (PRV)

    - The PRV will look like `[release version prefix].x`.
    - For example, `v0.32.1` is the first PRV for the v0.32.0 release.

## Steps

1. Create a branch from `main`.

2. Update the `CHANGELOG.md`.

    1. Modify CHANGELOG.md Header
        - First RCV or a PRV
            - Add a new header under `Main (unreleased)` for `[version] (YYYY-MM-DD)`.
        - Additional RCV or SRV
            - Update the header `[previous RCV] (YYYY-MM-DD)` to `[version] (YYYY-MM-DD)`.

    2. Move the unreleased changes included in the release branch from `Main (unreleased)` to `[version] (YYYY-MM-DD)`.

    3. Update **appropriate** places in the codebase that have the previous version with the new version determined above.
    
        *This will require some tribal knowledge not documented here (yet).*

        NOTE: Please do not update the operations/helm directory. It is updated independently from Agent releases for now.

3. Create a PR to merge to main (must be merged before continuing).

    See [here](https://github.com/grafana/agent/pull/2838/files) for an example PR for the first RCV.

    See [here](https://github.com/grafana/agent/pull/2873/files) for an example PR for a SRV.

4. Create a branch from `release-[release version prefix]`.

5. Cherry pick the change commit from the merged PR in step 3 from main into the new branch from step 4.

6. Create a PR to merge to `release-[release version prefix]` (must be merged before continuing).