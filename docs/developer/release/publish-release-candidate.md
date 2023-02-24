# Publish Release

## Steps

1. After a tag has been created during [Tag Release](./tag-release.md), GitHub 
Actions will create release assets and open a release draft for every pushed tag.

    - You can monitor this by viewing the drone build on the commit for the release tag.
    - This will take ~20-40 minutes.

2. You should see a new draft release created [here](https://github.com/grafana/agent/releases).

3. Edit the release draft by filling in the `Notable Changes` section with all `Breaking Changes` and `Feature` from the CHANGELOG.md.

4. Add any additional changes that you think are notable to the list.

5. Add a footer to the `Notable Changes` section `For a full list of changes, please refer to the [CHANGELOG](https://github.com/grafana/agent/blob/[version]/CHANGELOG.md)!`

6. At the bottom of the release page, tick the check box to "add a discussion" 
under the category for "announcements".

7. Also tick other boxes at the bottom of the release page:

    - For release candidates, tick the checkbox to "set as pre-release".
    - For stable releases and patch releases to the latest release branch, 
      tick the checkbox to "set as the latest release".

8. Optionally, have other team members review the release draft if you wish
   to feel more comfortable with it.

9. Publish the release!

10. Accounce the release in the Grafana Labs Community #agent channel

    Example message:

    ```
    :grafana-agent: Grafana Agent v0.32.0-rc.0 is now available! :grafana-agent:
    Release: https://github.com/grafana/agent/releases/tag/v0.32.0-rc.0
    Full changelog: https://github.com/grafana/agent/blob/v0.32.0-rc.0/CHANGELOG.md
    We'll be publishing v0.32.0 on Tuesday, February 28 if we haven't heard about any major issues.
    ```