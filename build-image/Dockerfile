# syntax=docker/dockerfile:1.4

# NOTE: This Dockerfile can only be built using BuildKit. BuildKit is used by
# default when running `docker buildx build` or when DOCKER_BUILDKIT=1 is set
# in environment variables.

#
# Dependencies
#
# We retrieve many of our dependencies by using various smaller containers.
#

# Dependency: golangci-lint (for linting)
FROM alpine as golangci
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.50.1

# Dependency: docker (for building images)
FROM alpine:3.16 as docker
RUN apk add --no-cache docker-cli docker-cli-buildx

# Dependency: Go and Go dependencies
FROM golang:1.19.3-bullseye as golang

# Keep in sync with cmd/agent-operator/DEVELOPERS.md
ENV CONTROLLER_GEN_VERSION v0.8.0

RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@$CONTROLLER_GEN_VERSION \
 && go install github.com/mitchellh/gox@v1.0.1                                         \
 && go install github.com/tcnksm/ghr@v0.15.0                                           \
 && go install github.com/grafana/tanka/cmd/tk@v0.22.1                                 \
 && go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@v0.5.1                \
 && go install github.com/google/go-jsonnet/cmd/jsonnet@v0.18.0                        \
 && go install github.com/golang/protobuf/protoc-gen-go@v1.3.1                         \
 && go install github.com/gogo/protobuf/protoc-gen-gogoslick@v1.3.0                    \
 && go install github.com/gogo/protobuf/gogoproto/...@v1.3.0                           \
 && go install github.com/ahmetb/gen-crd-api-reference-docs@v0.3.1-0.20220618162802-424739b250f5

#
# Final image
#

# rfratto/viceroy contains C cross compilers can be used for our Cgo
# dependencies.
FROM rfratto/viceroy:v0.2.1

# Install NodeJS LTS. This is needed because the most recent version of NodeJS
# from official Debian packages is v12, and we need LTS version v16.
#
# This must be done before installing other dependencies, otherwise nodesource
# will fail on installing NodeJS for all platforms instead of just our host
# platform.
RUN <<EOF
  apt-get update && apt-get install -qy curl
  curl -fsSL https://deb.nodesource.com/setup_16.x | bash -

  apt-get update && apt-get install -qy nodejs
  rm -rf /var/lib/apt/lists/*
EOF

# Install Yarn.
#
# See https://yarnpkg.com/getting-started/install#nodejs-1610
RUN corepack enable

# Install other dependencies.
#
# NOTE(rfratto): musl is installed so the Docker binaries from alpine work
# properly.
RUN apt-get update                                \
 && apt-get install -qy                           \
      build-essential file zip unzip gettext git  \
      musl libsystemd-dev nsis                    \
      rpm ruby ruby-dev rubygems                  \
      protobuf-compiler libprotobuf-dev           \
 && gem install --no-document fpm                 \
 && rm -rf /var/lib/apt/lists/*

COPY --from=golangci /bin/golangci-lint              /usr/local/bin
COPY --from=docker   /usr/bin/docker                 /usr/bin/docker
COPY --from=docker   /usr/libexec/docker/cli-plugins /usr/libexec/docker/cli-plugins
COPY --from=golang   /usr/local/go                   /usr/local/go
COPY --from=golang   /go/bin                         /go/bin

# Set CC to viceroycc to ensure that the cross compilers are used for all C
# compilation.
ENV CC viceroycc

ENV GOPATH /go
ENV PATH /usr/local/go/bin:/go/bin:$PATH
