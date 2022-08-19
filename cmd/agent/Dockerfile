FROM node:16.17-alpine as build-ui

# Enable Yarn
RUN corepack enable

COPY . /src/agent
WORKDIR /src/agent/web/ui
RUN yarn
RUN yarn build

FROM golang:1.18.0-bullseye as build
ARG RELEASE_BUILD=true
ARG IMAGE_TAG

# Backports repo required to get a libsystemd version 246 or newer which is required to handle journal +ZSTD compression
RUN echo "deb http://deb.debian.org/debian bullseye-backports main" >> /etc/apt/sources.list
RUN apt-get update && apt-get install -t bullseye-backports -qy libsystemd-dev

# Add support for bcc bindings required to compile the eBPF integration
RUN apt-get update && apt-get install -qy libbpfcc-dev

COPY . /src/agent
WORKDIR /src/agent

COPY --from=build-ui /src/agent/web/ui/build /src/agent/web/ui/build
RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false BUILTIN_ASSETS=true agent

FROM debian:bullseye-slim

# Backports repo required to get a libsystemd version 246 or newer which is required to handle journal +ZSTD compression
# plus the libbpfcc library for running the eBPF integration.
RUN echo "deb http://deb.debian.org/debian bullseye-backports main" >> /etc/apt/sources.list
RUN apt-get update && apt-get install -t bullseye-backports -qy libsystemd-dev && \
  apt-get install -qy tzdata ca-certificates && \
  if [ `uname -m` = "x86_64" ]; then apt-get install -qy libbpfcc; fi && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build /src/agent/cmd/agent/agent /bin/agent
COPY cmd/agent/agent-local-config.yaml /etc/agent/agent.yaml

ENTRYPOINT ["/bin/agent"]
CMD ["--config.file=/etc/agent/agent.yaml", "--metrics.wal-directory=/etc/agent/data"]
