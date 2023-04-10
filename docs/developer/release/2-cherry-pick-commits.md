# Cherry Pick Commits

Any commits not on the release branch need to be cherry-picked over to it. If the release branch
has all code changes on it, skip this step.

## Steps

1. Create PR(s) to cherry-pick additional commits into the release branch as needed from main:

    ```
    git cherry-pick -x COMMIT_SHA
    ```

    - For example, refer to PR [#3188](https://github.com/grafana/agent/pull/3188) and [#3185](https://github.com/grafana/agent/pull/3185). 
