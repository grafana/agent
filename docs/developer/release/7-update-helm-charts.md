# Update Release Branch

Our helm charts require some version updates as well.

## Steps

1. Create a branch from `main` for [grafana/helm-charts](https://github.com/grafana/helm-charts).

2. Update the code.

    1. Copy the content of the last CRDs into helm-charts.
        
        Copy the contents from agent repo `production/operator/crds/` to replace the contents of helm-charts repo `charts/agent-operator/crds`
        
    2. Update references of agent-operator app version in helm-charts pointing to release version.
    3. Bump up the helm chart version.
    
    *NOTE: There is no need to update the README.md manually - running the 
[helm-docs](https://github.com/norwoodj/helm-docs) utility in the `charts/agent-operator` directory 
will update it automatically.*

2. Open a PR.

    - Example PR [here](https://github.com/grafana/helm-charts/pull/2233)

3. Create a branch from `main` for [grafana/agent](https://github.com/grafana/agent).

4. Update the code.
    
    1. Update `Chart.yaml` with the new helm version and app version
    2. Update `CHANGELOG.md` with a new section for the helm version
    3. Run `make generate-helm-docs generate-helm-tests` from the repo root

5. Open a PR.

    - PRV example PR [here](https://github.com/grafana/agent/pull/3126)