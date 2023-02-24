# Create Release Branch

This is the first step taken as release shepherd for a new release 
after [Prerequisites](./prerequisites.md) are met.

A release branch is created for every major or minor release. That release
branch is then used for all Release Candidates, the Stable Release, and all
Patch Releases.

## Steps

1. Gather consensus on which commit should be used as a base for the release
   branch.

2. Determine the version and prefix.

    The release version prefix can be determined by looking at the last version and adding to it. 

    - v0.31.0 -> v0.32
    - v0.31.3 -> v0.32

    An exception could occur here for a major release.

    - v0.10.0 -> v1.0
    - v0.31.0 -> v1.0

3. Create and push the release branch from the selected base commit:

    The name of the release branch should be `release-` suffixed with the 
    version prefix defined in step 2, such as `release-v0.32`.

        Note: There is no branch such as `release-v0.32-rc.0` or `release-v0.32.0`.

    - If the consensus commit is the latest commit from main you can branch from main.
    - If the consensus commit is not the latest commit from main, branch from that instead.

4. Create a PR to cherry-pick additional commits into the release branch as
   needed. 