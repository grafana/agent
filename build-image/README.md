# Grafana Agent build images

The Grafana Agent build images are used for CI workflows to manage builds of
Grafana Agent.

There are three [images][agent-build-image-dockerhub]:

* `grafana/agent-build-image:X.Y.Z` (for building targeting Linux, including Linux boringcrypto)
* `grafana/agent-build-image:X.Y.Z-windows` (for builds targeting Windows)
* `grafana/agent-build-image:X.Y.Z-boringcrypto` (for building targeting Windows boringcrypto)

(Where `X.Y.Z` is replaced with some semantic version, like 0.14.0).

[agent-build-image-dockerhub]:https://hub.docker.com/repository/docker/grafana/agent-build-image/general

## Creating new images

### Step 1: Update the main branch

Open a PR to update the build images. 
See [this][example-pr] pull request for an example.
You need to change the following files:
 * `build-image/Dockerfile`
 * `build-image/windows/Dockerfile`
 * `.drone/drone.yaml`
 * `.drone/pipelines/build_images.jsonnet`
 * `.github/workflows/check-linux-build-image.yml`

[example-pr]:https://github.com/grafana/agent/pull/6650/files

### Step 2: Create a Git tag

After the PR is merged to `main`, a maintainer must push a tag matching the pattern 
`build-image/vX.Y.Z` to the `grafana/agent` repo. 
For example, to create version `0.41.0` of the build images,
a maintainer would push the tag `build-image/v0.41.0`:

```
git checkout main
git pull
git tag -s build-image/v0.41.0
git push origin build-image/v0.41.0
```

> **NOTE**: The tag name is expected to be prefixed with `v`, but the pushed
> images have the `v` prefix removed.

> **NOTE**: The tag name doesn't have to correspond to an Agent version.

Automation will trigger off of this tag being pushed, building and pushing the
new build images to Docker Hub.

A follow-up commit to use the newly pushed build images must be made.
