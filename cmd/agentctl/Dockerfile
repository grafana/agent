FROM golang:1.15 as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=false
ARG IMAGE_TAG
RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agentctl

FROM debian:buster-slim
RUN apt-get update && \
  apt-get install -qy \
  tzdata ca-certificates && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
COPY --from=build /src/agent/cmd/agentctl/agentctl /bin/agentctl

ENTRYPOINT ["/bin/agentctl"]
