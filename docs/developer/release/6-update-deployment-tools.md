# Update Deployment Tools

We must update [grafana/deployment_tools](https://github.com/grafana/deployment_tools) (unless it is already in the desired state) to test our RCV or to stop testing it.

## Steps

1. Create a branch from `main` for [grafana/deployment_tools](https://github.com/grafana/deployment_tools).

2. Open a PR to update Update [main.jsonnet](https://github.com/grafana/deployment_tools/blob/master/ksonnet/environments/grafana-agent/main.jsonnet).

    - RCV example PR [here](https://github.com/grafana/deployment_tools/pull/58203)
    - SRV example PR [here](https://github.com/grafana/deployment_tools/pull/58674)

    Wait for the build to complete after merge.

3. Validate the new version has been deployed to dev_canary.

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

4. Validate the new version is healthy in dev_canary

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