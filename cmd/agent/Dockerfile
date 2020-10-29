FROM golang:1.15 as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=true
ARG IMAGE_TAG
RUN apt-get update && apt-get install -qy libsystemd-dev
RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agent

FROM debian:buster-slim
RUN apt-get update && \
  apt-get install -qy \
  tzdata ca-certificates libsystemd-dev && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
COPY --from=build /src/agent/cmd/agent/agent /bin/agent
COPY cmd/agent/agent-local-config.yaml /etc/agent/agent.yaml

ENTRYPOINT ["/bin/agent"]
CMD ["--config.file=/etc/agent/agent.yaml", "--prometheus.wal-directory=/etc/agent/data"]
