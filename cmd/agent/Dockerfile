FROM golang:1.13 as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=false
RUN make clean && make RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agent

FROM alpine:3.9
RUN apk add --update --no-cache ca-certificates
COPY --from=build /src/agent/cmd/agent/agent /bin/agent
COPY cmd/agent/agent-local-config.yaml /etc/agent/local-config.yaml

ENTRYPOINT ["/bin/agent"]
CMD ["-config.file=/etc/agent/local-config.yaml"]
