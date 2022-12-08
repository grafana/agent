# syntax=docker/dockerfile:1.4

# NOTE: This Dockerfile can only be built using BuildKit. BuildKit is used by
# default when running `docker buildx build` or when DOCKER_BUILDKIT=1 is set
# in environment variables.

FROM --platform=$BUILDPLATFORM grafana/agent-build-image:0.17.0 as build
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG RELEASE_BUILD=1
ARG VERSION

COPY . /src/agent
WORKDIR /src/agent

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    RELEASE_BUILD=${RELEASE_BUILD} VERSION=${VERSION} \
    make agentctl

FROM ubuntu:jammy

# Install dependencies needed at runtime.
RUN <<EOF
  apt-get update
  apt-get install -qy libsystemd-dev tzdata ca-certificates

  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
EOF

COPY --from=build /src/agent/build/agentctl /bin/agentctl
ENTRYPOINT ["/bin/agentctl"]
