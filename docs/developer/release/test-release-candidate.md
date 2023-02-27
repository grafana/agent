# Test Release Candidate

This Action will typically follow [Publish Release Candidate](./publish-release-candidate.md).

During this action we will deploy the new Release Candidate Version (RCV) to dev_canary. This
will involve overriding the deployed default version to point to the RCV.

## Steps

1. Create a branch from `main` for [grafana/deployment_tools](https://github.com/grafana/deployment_tools).

2. Update [main.jsonnet](https://github.com/grafana/deployment_tools/blob/master/ksonnet/environments/grafana-agent/main.jsonnet)

    ```
    local agentWaves = (import 'waves/agent.libsonnet'),
    ```

    becomes

    ```
    local agentWaves = (import 'waves/agent.libsonnet') {
        // Temporary override while testing v0.32.0-rc.0
        dev_canary: 'grafana/agent:v0.32.0-rc.0',
    },
    ```

    Example PR: https://github.com/grafana/deployment_tools/pull/58203

3. Open a PR and wait for it to be merged.

4. Validate the new version has been deployed to dev_canary

    *TODO*