# Create Release Branch

This is the first step taken as release shephard after [Prerequisites](./prerequisites.md) are met. This branch will be used for the Release Candidates, Stable Release and Patch Releases.

1. Gather consensus on which commit should be used as a base for the release
   branch.

2. Determine the new minor version prefix

    The release version prefix can be determined by looking at the last version and adding to it. 

    - v0.31.0 -> v0.32
    - v0.31.3 -> v0.32

    An exception could occur here for a major release.

    - v0.10.0 -> v1.0
    - v0.31.0 -> v1.0

3. Create and push the release branch from the selected base commit:

    The name of the release branch should be the name of the stable version we intend 
    to release prefixed with the release version prefix determined in step 2, such as `release-v0.31`.

    Note: There is no branch such as `release-v0.31-rc.0` or `release-v0.31.0`. Only use the version prefix as shown above.

    If the connsensus commit is the latest commit from main:

    ```mermaid
    gitGraph
        commit id: "1"
        commit id: "2"
        commit id: "3"
        branch release-v0.31
        checkout release-v0.31
        cherry-pick id: "3"
    ```

    If the consensus commit is not the latest commit from main:

    ```mermaid
    gitGraph
        commit id: "1"
        branch release-v0.31
        checkout release-v0.31
        cherry-pick id: "1"
        checkout main
        commit id: "2"
        commit id: "3"
    ```

4. Create a PR to cherry-pick additional commits into the release branch as
   needed. 

    ```mermaid
    gitGraph
        commit id: "1"
        commit id: "2"
        commit id: "3"
        branch release-v0.31
        checkout release-v0.31
        cherry-pick id: "3"
        checkout main
        commit id: "4"
        commit id: "5"
        checkout release-v0.31
        cherry-pick id: "5"
    ```