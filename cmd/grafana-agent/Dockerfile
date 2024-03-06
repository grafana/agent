# syntax=docker/dockerfile:1.4

# NOTE: This Dockerfile can only be built using BuildKit. BuildKit is used by
# default when running `docker buildx build` or when DOCKER_BUILDKIT=1 is set
# in environment variables.

FROM --platform=$BUILDPLATFORM grafana/agent-build-image:0.33.0 as build
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG RELEASE_BUILD=1
ARG VERSION
ARG GOEXPERIMENT

COPY . /src/agent
WORKDIR /src/agent

# Build the UI before building the agent, which will then bake the final UI
# into the binary.
RUN --mount=type=cache,target=/src/agent/web/ui/node_modules,sharing=locked \
   make generate-ui

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    RELEASE_BUILD=${RELEASE_BUILD} VERSION=${VERSION} \
    GO_TAGS="netgo builtinassets promtail_journal_enabled" \
    GOEXPERIMENT=${GOEXPERIMENT} \
    make agent

FROM public.ecr.aws/ubuntu/ubuntu:mantic

LABEL org.opencontainers.image.source="https://github.com/grafana/agent"

# Install dependencies needed at runtime.
RUN <<EOF
  apt-get update
  apt-get install -qy libsystemd-dev tzdata ca-certificates
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
EOF

COPY --from=build /src/agent/build/grafana-agent /bin/grafana-agent
COPY cmd/grafana-agent/agent-local-config.yaml /etc/agent/agent.yaml

ARG UID=0
ARG GF_UID=473
ARG GF_UNAME="grafana-agent"

RUN groupadd --gid $GF_UID $GF_UNAME
RUN useradd -m -u $GF_UID -g $GF_UID $GF_UNAME
RUN chown -R $GF_UNAME:$GF_UNAME /etc/agent
RUN chown -R $GF_UNAME:$GF_UNAME /bin/grafana-agent
USER $UID:$UID

ENTRYPOINT ["/bin/grafana-agent"]
ENV AGENT_DEPLOY_MODE=docker
CMD ["--config.file=/etc/agent/agent.yaml", "--metrics.wal-directory=/etc/agent/data"]
