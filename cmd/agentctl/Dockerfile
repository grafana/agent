FROM golang:1.14 as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=false
ARG IMAGE_TAG
RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agentctl

FROM alpine:3.9
RUN apk add --update --no-cache ca-certificates
COPY --from=build /src/agent/cmd/agentctl/agentctl /bin/agentctl

ENTRYPOINT ["/bin/agentctl"]
