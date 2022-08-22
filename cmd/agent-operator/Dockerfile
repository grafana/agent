FROM golang:1.18.0-bullseye as build
COPY . /src/agent
WORKDIR /src/agent
ARG RELEASE_BUILD=false
ARG IMAGE_TAG

RUN make clean && make IMAGE_TAG=${IMAGE_TAG} RELEASE_BUILD=${RELEASE_BUILD} BUILD_IN_CONTAINER=false agent-operator

FROM gcr.io/distroless/static-debian11

COPY --from=build /src/agent/cmd/agent-operator/agent-operator /bin/agent-operator

ENTRYPOINT ["/bin/agent-operator"]
