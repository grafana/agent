# Test Release Candidate

This Action will typically follow [Publish Release Candidate](./publish-release-candidate.md).

During this action we will deploy the new Release Candidate Version (RCV) to dev_canary. This
will involve overriding the deployed default version to point to the RCV.

## Steps

1. Create a branch from `main` for [grafana/deployment_tools](https://github.com/grafana/deployment_tools).

2. Update [main.jsonnet](https://github.com/grafana/deployment_tools/blob/master/ksonnet/environments/grafana-agent/main.jsonnet).

    ```
    local agentWaves = (import 'waves/agent.libsonnet'),
    ```

    becomes

    ```
    local agentWaves = (import 'waves/agent.libsonnet') {
        // Temporary override while testing [release version]
        dev_canary: 'grafana/agent:[release version]',
    },
    ```

    Example PR: https://github.com/grafana/deployment_tools/pull/58203

3. Open a PR and wait for it to be merged.

    Wait for the build to complete after merge.

4. Validate the new version has been deployed to dev_canary.

    1. Log in to the [Grafana Admin Home](https://admin-dev-us-central-0.grafana.net/grafana/?orgId=1).

    2. Navigate to Explore in the left hand panel.

    3. Run the following query to make sure there are hits for the [release version].

    ```
    agent_build_info{version="[release version]",cluster="dev-us-central-0",namespace="grafana-agent"}
    ```

    *NOTE: The datasource should be set to `cortex-dev-01-dev-us-central-0`*.

    4. Run the following query to make sure there are 0 hits for any other version.

    ```
    agent_build_info{version!="[release version]",cluster="dev-us-central-0",namespace="grafana-agent"}
    ```

5. Validate the new version is healthy in dev_canary

    1. Do steps 2-4 for the following dashboards.
        - `Grafana Agent Flow / Controller Dashboard`
        - `Grafana Agent Flow / Resources`
        - `Grafana Agent Flow / prometheus.remote_write`

    2. Select the Filters.
        - `Data Source` = `cortex-dev-01-dev-us-central-0`
        - `Loki Data Source` = `loki-dev`
        - `cluster` = `dev-us-central-0`
        - `namespace` = `grafana-agent`

    3. Make sure all components are healthy.

    4. Review the graphs for a time period before and after the new version started running to make sure nothing sticks out.