# Update Homebrew

After a stable or patch release is created, a bot will automatically create a PR in the [homebrew-grafana][] repository.
The PR will bump the version of Agent in Agent's Brew formula.

There will only be one PR for each release, and it will be for `grafana-agent-flow`.
There is no Brew formula for `grafana-agent`. 

## Steps

1. Navigate to the [homebrew-grafana][] repository.

2. Find the PR which bumps the Agent formula to the release that was just published. It will look like [this one][example-pr].

3. Merge the PR.

[homebrew-grafana]: https://github.com/grafana/homebrew-grafana
[example-pr]: https://github.com/grafana/homebrew-grafana/pull/87