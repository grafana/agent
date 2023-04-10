# Create Release Branch

A single release branch is created for every major or minor release. That release
branch is then used for all Release Candidates, the Stable Release, and all
Patch Releases.

## Release Version Prefix

The `[release version prefix]` can be determined by looking at the last release
version and adding to it. Also, drop off the 3rd number as shown in the examples below.

- v0.30.0 -> v0.31
- v0.30.3 -> v0.31

For a major version we jump up the first number as shown in the examples below.

- v0.10.0 -> v1.0
- v0.31.0 -> v1.0

*NOTE: This value will be referred to as `[release version prefix]` in this documentation*

## Steps

1. Determine which commit should be used as a base for the release branch.

2. Create and push the release branch from the selected base commit:

    The name of the release branch should be `release-[release version prefix]`
    defined above, such as `release-v0.31`.

        *NOTE: There is never a branch such as `release-v0.31-rc.0` or `release-v0.31.0`.*

    - If the consensus commit is the latest commit from main you can branch from main.
    - If the consensus commit is not the latest commit from main, branch from that instead.

    *NOTE: don't create any other branches that are prefixed with `release` when creating PRs or
    those branches will collide with our automated release build publish rules.*