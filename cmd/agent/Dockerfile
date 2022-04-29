FROM golang:1.18.0-bullseye as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=true
ARG IMAGE_TAG

# Backports repo required to get a libsystemd version 246 or newer which is required to handle journal +ZSTD compression
RUN echo "deb http://deb.debian.org/debian bullseye-backports main" >> /etc/apt/sources.list
RUN apt-get update && apt-get install -t bullseye-backports -qy libsystemd-dev

# Add support for bcc bindings required to compile the eBPF integration
RUN echo "deb [trusted=yes] https://repo.iovisor.org/apt/bionic bionic-nightly main" | tee /etc/apt/sources.list.d/iovisor.list
RUN apt-get update && apt-get install -qy libbpfcc-dev

RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agent

FROM debian:bullseye-slim

# Backports repo required to get a libsystemd version 246 or newer which is required to handle journal +ZSTD compression
RUN echo "deb http://deb.debian.org/debian bullseye-backports main" >> /etc/apt/sources.list
RUN apt-get update && apt-get install -t bullseye-backports -qy libsystemd-dev && \
  apt-get install -qy tzdata ca-certificates && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build /src/agent/cmd/agent/agent /bin/agent
COPY cmd/agent/agent-local-config.yaml /etc/agent/agent.yaml

ENTRYPOINT ["/bin/agent"]
CMD ["--config.file=/etc/agent/agent.yaml", "--metrics.wal-directory=/etc/agent/data"]
