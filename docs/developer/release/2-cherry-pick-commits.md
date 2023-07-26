# Cherry Pick Commits

Any commits not on the release branch need to be cherry-picked over to it.

## Before you begin

1. If the release branch already has all code changes on it, skip this step.

## Steps

1. Create PR(s) to cherry-pick additional commits into the release branch as needed from main:

    ```
    git cherry-pick -x COMMIT_SHA
    ```
    - If there are several commits to cherry-pick, consider using one branch to cherry-pick to and ask reviewers to review commit-by-commit.

    - For example, refer to PR [#3188](https://github.com/grafana/agent/pull/3188) and [#3185](https://github.com/grafana/agent/pull/3185). 
