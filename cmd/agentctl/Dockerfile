FROM golang:1.18.0-bullseye as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=false
ARG IMAGE_TAG

# Backports repo required to get a libsystemd version 246 or newer which is required to handle journal +ZSTD compression
RUN echo "deb http://deb.debian.org/debian bullseye-backports main" >> /etc/apt/sources.list
RUN apt-get update && apt-get install -t bullseye-backports -qy libsystemd-dev

# Add support for bcc bindings required to compile the eBPF integration
RUN apt-get update && apt-get install -qy libbpfcc-dev

RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agentctl

FROM ubuntu:jammy

RUN apt-get update && apt-get install -qy libsystemd-dev && \
  apt-get install -qy tzdata ca-certificates && \
  if [ `uname -m` = "x86_64" ]; then apt-get install -qy libbpfcc; fi && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build /src/agent/cmd/agentctl/agentctl /bin/agentctl

ENTRYPOINT ["/bin/agentctl"]
