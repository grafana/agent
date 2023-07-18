# Update Helm Charts

Our Helm charts require some version updates as well.

## Before you begin

1. Install [helm-docs](https://github.com/norwoodj/helm-docs) on macOS/Linux.

## Steps

1. Create a branch from `main` for [grafana/agent](https://github.com/grafana/agent).

2. Update the code:
    
    1. Update `Chart.yaml` with the new helm version and app version.
    2. Update `CHANGELOG.md` with a new section for the helm version.
    3. Run `make docs rebuild-tests` from the repo root.

3. Open a PR, following the pattern in PR [#3126](https://github.com/grafana/agent/3126).