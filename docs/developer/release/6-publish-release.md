# Publish Release Candidate

This is how to publish the release in GitHub.

## Steps

1. You should see a new draft release created [here](https://github.com/grafana/agent/releases). If not go back to [Tag Release](./4-tag-release.md).

2. Edit the release draft by filling in the `Notable Changes` section with all `Breaking Changes` and `Features` from the CHANGELOG.md.

3. Add any additional changes that you think are notable to the list.

4. Add a footer to the `Notable Changes` section:

    `For a full list of changes, please refer to the [CHANGELOG](https://github.com/grafana/agent/blob/RELEASE_VERSION/CHANGELOG.md)!`
    
    Do not substitute the value for `CHANGELOG`.  

5. At the bottom of the release page, perform the following:
    - Tick the check box to "add a discussion" under the category for "announcements".
    - For a RCV, tick the checkbox to "pre-release".
    - For a SRV or PRV, tick the checkbox to "set as the latest release".

6. Optionally, have other team members review the release draft if you wish
   to feel more comfortable with it.

7. Publish the release!