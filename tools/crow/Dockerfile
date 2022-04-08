FROM golang:1.18.0-bullseye as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=true
ARG IMAGE_TAG

RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false grafana-agent-crow
RUN apt update && apt install ca-certificates

FROM debian:bullseye-slim
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /src/agent/tools/crow/grafana-agent-crow /bin/grafana-agent-crow
ENTRYPOINT ["/bin/grafana-agent-crow"]
