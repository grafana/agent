FROM golang:1.16-buster as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=true
ARG IMAGE_TAG

RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false grafana-agent-crow

FROM debian:buster-slim
COPY --from=build /src/agent/tools/crow/grafana-agent-crow /bin/grafana-agent-crow
ENTRYPOINT ["/bin/grafana-agent-crow"]
