# Update Version in Code

The project must be updated to reference the upcoming release tag whenever a new release is being made.

## Release Version

The release version prefix was previously determined in [Create Release Branch](./1-create-release-branch.md). The `[release version]` is defined as follows.

- Release Candidate Version (RCV)

    - The RCV will look like `[release version prefix].x-rc.y`.
    - For example, `v0.31.0-rc.0` is the first RCV for the v0.31.0 release.

- Stable Release Version (SRV)

    - The SRV will look like `[release version prefix].0`.
    - For example, `v0.31.0` is the SRV for the v0.31.0 release.

- Patch Release Version (PRV)

    - The PRV will look like `[release version prefix].x`.
    - For example, `v0.31.1` is the first PRV for the v0.31.0 release.

*NOTE: This value will be referred to as `[release version]` in this documentation*

## Steps

1. Create a branch from `main` for [grafana/agent](https://github.com/grafana/agent).

2. Update the `CHANGELOG.md`.

    1. `CHANGELOG.md` Header
        - First RCV or a PRV
            - Add a new header under `Main (unreleased)` for `[release version] (YYYY-MM-DD)`.
        - Additional RCV or SRV
            - Update the header `[previous RCV] (YYYY-MM-DD)` to `[release version] (YYYY-MM-DD)`. The date may need updating.

    2. Move the unreleased changes we want to add to the release branch from `Main (unreleased)` to `[release version] (YYYY-MM-DD)`.

    3. Update appropriate places in the codebase that have the previous version with the new version determined above.

        *NOTE: Please do not update the operations/helm directory. It is updated independently from Agent releases for now.*
    
        *NOTE: This will require some tribal knowledge not documented here (yet).*

3. Create a PR to merge to main (must be merged before continuing).

    - RCV example PR [here](https://github.com/grafana/agent/pull/3065)
    - SRV example PR [here](https://github.com/grafana/agent/pull/3119)
    - PRV example PR [here](https://github.com/grafana/agent/pull/3191)

4. Create a branch from `release-[release version prefix]` for [grafana/agent](https://github.com/grafana/agent).

5. Cherry pick the commit on main from the merged PR in Step 3 from main into the new branch from Step 4.

    ```
    git cherry-pick -x [commit id]
    ```

    For a SRV, delete the `Main (unreleased)` header and anything underneath it as part of the cherry-pick. Alternatively, do it after the cherry-pick is completed.

6. Create a PR to merge to `release-[release version prefix]` (must be merged before continuing).

    - RCV example PR [here](https://github.com/grafana/agent/pull/3066)
    - SRV example PR [here](https://github.com/grafana/agent/pull/3123)
    - PRV example PR [here](https://github.com/grafana/agent/pull/3193)
        - The `CHANGELOG.md` was updated in cherry-pick commits prior for this example. Make sure it is all set on this PR.