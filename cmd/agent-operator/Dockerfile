FROM golang:1.18.5-alpine3.16 as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=false
ARG IMAGE_TAG

RUN apk add --update make bash git

RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agent-operator

FROM alpine:3.16

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=build /src/agent/cmd/agent-operator/agent-operator /bin/agent-operator

ENTRYPOINT ["/bin/agent-operator"]
