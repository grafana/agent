## Build, test, and generate code for various parts of Grafana Agent.
##
## At least Go 1.18, git, and a moderately recent version of Docker is required
## to be able to use the Makefile. If you do not have the full list of build
## dependencies, you may set USE_CONTAINER=1 to proxy build commands to a build
## container.
##
## Other environment variables can be used to tweak behaviors of targets.
## See the bottom of this help section for the full list of supported
## environment variables.
##
## Usage:
##   make <target>
##
## Targets for running tests:
##
##   test  Run tests
##   lint  Lint code
##
## Targets for building binaries:
##
##   binaries  Compiles all binaries.
##   agent     Compiles cmd/agent to $(AGENT_BINARY)
##   agentctl  Compiles cmd/agentctl to $(AGENTCTL_BINARY)
##   operator  Compiles cmd/agent-operator to $(OPERATOR_BINARY)
##   crow      Compiles tools/crow to $(CROW_BINARY)
##   smoke     Compiles tools/smoke to $(SMOKE_BINARY)
##
## Targets for building Docker images:
##
##   images          Builds all Docker images.
##   agent-image     Builds agent Docker image.
##   agentctl-image  Builds agentctl Docker image.
##   operator-image  Builds operator Docker image.
##   crow-image      Builds crow Docker image.
##   smoke-image     Builds smoke test Docker image.
##
## Targets for generating assets:
##
##   generate             Generate everything.
##   generate-crds        Generate Grafana Agent Operator CRDs.
##   generate-manifests   Generate production/kubernetes YAML manifests.
##   generate-dashboards  Generate dashboards in example/docker-compose after
##                        changing Jsonnet.
##   generate-protos      Generate protobuf files.
##
## Other targets:
##
##   build-container-cache  Create a cache for the build container to speed up
##                          subsequent proxied builds
##   clean                  Clean caches and built binaries
##   help                   Displays this message
##   info                   Print Makefile-specific environment variables
##
## Environment variables:
##
##   USE_CONTAINER    Set to 1 to enable proxying commands to build container
##   AGENT_IMAGE      Image name:tag built by `make agent-image`
##   AGENTCTL_IMAGE   Image name:tag built by `make agentctl-image`
##   OPERATOR_IMAGE   Image name:tag built by `make operator-image`
##   CROW_IMAGE       Image name:tag built by `make crow-image`
##   SMOKE_IMAGE      Image name:tag built by `make smoke-image`
##   BUILD_IMAGE      Image name:tag used by USE_CONTAINER=1
##   AGENT_BINARY     Output path of `make agent` (default build/agent)
##   AGENTCTL_BINARY  Output path of `make agentctl` (default build/agentctl)
##   OPERATOR_BINARY  Output path of `make operator` (default build/agent-operator)
##   CROW_BINARY      Output path of `make crow` (default build/agent-crow)
##   SMOKE_BINARY     Output path of `make smoke` (default build/agent-smoke)
##   GOOS             Override OS to build binaries for
##   GOARCH           Override target architecture to build binaries for
##   GOARM            Override ARM version (6 or 7) when GOARCH=arm
##   CGO_ENABLED      Set to 0 to disable Cgo for builds
##   RELEASE_BUILD    Set to 1 to build release binaries
##   VERSION          Version to inject into built binaries.
##   GO_TAGS          Extra tags to use when building.

# TODO(rfratto): test-packages target
#
# This depends on some reworking of how the Go tests in ./packaging works to
# assume that packages have already been built so we don't have to hook in
# packaging code here.

include tools/make/*.mk

AGENT_IMAGE     ?= grafana/agent:latest
AGENTCTL_IMAGE  ?= grafana/agentctl:latest
OPERATOR_IMAGE  ?= grafana/agent-operator:latest
CROW_IMAGE      ?= us.gcr.io/kubernetes-dev/grafana/agent-crow:latest
SMOKE_IMAGE     ?= us.gcr.io/kubernetes-dev/grafana/agent-smoke:latest
AGENT_BINARY    ?= build/agent
AGENTCTL_BINARY ?= build/agentctl
OPERATOR_BINARY ?= build/agent-operator
CROW_BINARY     ?= build/agent-crow
SMOKE_BINARY    ?= build/agent-smoke
GOOS            ?= $(shell go env GOOS)
GOARCH          ?= $(shell go env GOARCH)
GOARM           ?= $(shell go env GOARM)
CGO_ENABLED     ?= 1
RELEASE_BUILD   ?= 0

# This should contain the list of all environment variables should should
# propagate to the build container. USE_CONTAINER should _not_ be included to
# avoid infinite recursion.
PROPAGATE_VARS := \
	AGENT_IMAGE AGENTCTL_IMAGE OPERATOR_IMAGE CROW_IMAGE SMOKE_IMAGE \
	BUILD_IMAGE GOOS GOARCH GOARM CGO_ENABLED RELEASE_BUILD \
	AGENT_BINARY AGENTCTL_BINARY OPERATOR_BINARY CROW_BINARY SMOKE_BINARY \
	VERSION GO_TAGS

#
# Contants for targets
#

GO_ENV := GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) CGO_ENABLED=$(CGO_ENABLED)

VERSION      ?= $(shell ./tools/image-tag)
GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_BRANCH   := $(shell git rev-parse --abbrev-ref HEAD)
VPREFIX      := github.com/grafana/agent/pkg/build
GO_LDFLAGS   := -X $(VPREFIX).Branch=$(GIT_BRANCH)                        \
                -X $(VPREFIX).Version=$(VERSION)                          \
                -X $(VPREFIX).Revision=$(GIT_REVISION)                    \
                -X $(VPREFIX).BuildUser=$(shell whoami)@$(shell hostname) \
                -X $(VPREFIX).BuildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

DEFAULT_FLAGS    := $(GO_FLAGS)
DEBUG_GO_FLAGS   := -gcflags "all=-N -l" -ldflags "$(GO_LDFLAGS)" -tags "netgo $(GO_TAGS)"
RELEASE_GO_FLAGS := -ldflags "-s -w $(GO_LDFLAGS)" -tags "netgo $(GO_TAGS)"

ifeq ($(RELEASE_BUILD),1)
GO_FLAGS := $(DEFAULT_FLAGS) $(RELEASE_GO_FLAGS)
else
GO_FLAGS := $(DEFAULT_FLAGS) $(DEBUG_GO_FLAGS)
endif

#
# Targets for running tests
#
# These targets currently don't support proxying to a build container due to
# difficulties with testing ./pkg/util/k8s and testing packages.
#

.PHONY: lint
lint:
	golangci-lint run -v --timeout=10m

.PHONY: test
# We have to run test twice: once for all packages with -race and then once
# more without -race for packages that have known race detection issues.
test:
	$(GO_ENV) go test $(CGO_FLAGS) -race ./...
	$(GO_ENV) go test $(CGO_FLAGS) ./pkg/integrations/node_exporter ./pkg/logs ./pkg/operator ./pkg/util/k8s

#
# Targets for building binaries
#

.PHONY: binaries agent agentctl operator crow smoke
binaries: agent agentctl operator crow smoke

agent:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(AGENT_BINARY) ./cmd/agent
endif

agentctl:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(AGENTCTL_BINARY) ./cmd/agentctl
endif

operator:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(OPERATOR_BINARY) ./cmd/agent-operator
endif

crow:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(CROW_BINARY) ./tools/crow
endif

smoke:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(SMOKE_BINARY) ./tools/smoke
endif

#
# Targets for building Docker images
#

.PHONY: images agent-image agentctl-image operator-image crow-image smoke-image
images: agent-image agentctl-image operator-image crow-image smoke-image

agent-image: GOOS         := linux
agent-image: AGENT_BINARY := build/docker/agent
agent-image: agent
	docker build -t $(AGENT_IMAGE) -f cmd/agent/Dockerfile .

agentctl-image: GOOS            := linux
agentctl-image: AGENTCTL_BINARY := build/docker/agentctl
agentctl-image: agentctl
	docker build -t $(AGENTCTL_IMAGE) -f cmd/agentctl/Dockerfile .

operator-image: GOOS            := linux
operator-image: OPERATOR_BINARY := build/docker/agent-operator
operator-image: operator
	docker build -t $(OPERATOR_IMAGE) -f cmd/agent-operator/Dockerfile .

crow-image: GOOS        := linux
crow-image: CROW_BINARY := build/docker/agent-crow
crow-image: crow
	docker build -t $(CROW_IMAGE) -f tools/crow/Dockerfile .

smoke-image: GOOS         := linux
smoke-image: SMOKE_BINARY := build/docker/agent-smoke
smoke-image: smoke
	docker build -t $(SMOKE_IMAGE) -f tools/smoke/Dockerfile .

#
# Targets for generating assets
#

.PHONY: generate generate-crds generate-manifests generate-dashboards generate-protos
generate: generate-crds generate-manifests generate-dashboards generate-protos

generate-crds:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	bash ./tools/generate-crds.bash
endif

generate-manifests:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	cd production/kubernetes/build && bash build.sh
endif

generate-dashboards:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	cd example/docker-compose && jb install && \
	cd grafana/dashboards && jsonnet template.jsonnet -J ../../vendor -m .
endif

generate-protos:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	go generate ./pkg/agentproto/
endif

#
# Other targets
#
# build-container-cache and clean-build-container-cache are defined in
# Makefile.build-container.
#

.PHONY: clean
clean: clean-build-container-cache
	rm -rf ./build/*
	rm -rf ./dist/*

.PHONY: info
info:
	@printf "USE_CONTAINER   = $(USE_CONTAINER)\n"
	@printf "AGENT_IMAGE     = $(AGENT_IMAGE)\n"
	@printf "AGENTCTL_IMAGE  = $(AGENTCTL_IMAGE)\n"
	@printf "OPERATOR_IMAGE  = $(OPERATOR_IMAGE)\n"
	@printf "CROW_IMAGE      = $(CROW_IMAGE)\n"
	@printf "SMOKE_IMAGE     = $(SMOKE_IMAGE)\n"
	@printf "BUILD_IMAGE     = $(BUILD_IMAGE)\n"
	@printf "AGENT_BINARY    = $(AGENT_BINARY)\n"
	@printf "AGENTCTL_BINARY = $(AGENTCTL_BINARY)\n"
	@printf "OPERATOR_BINARY = $(OPERATOR_BINARY)\n"
	@printf "CROW_BINARY     = $(CROW_BINARY)\n"
	@printf "SMOKE_BINARY    = $(SMOKE_BINARY)\n"
	@printf "GOOS            = $(GOOS)\n"
	@printf "GOARCH          = $(GOARCH)\n"
	@printf "GOARM           = $(GOARM)\n"
	@printf "CGO_ENABLED     = $(CGO_ENABLED)\n"
	@printf "RELEASE_BUILD   = $(RELEASE_BUILD)\n"
	@printf "VERSION         = $(VERSION)\n"
	@printf "GO_TAGS         = $(GO_TAGS)\n"

# awk magic to print out the comment block at the top of this file.
.PHONY: help
help:
	@awk 'BEGIN {FS="## "} /^##\s*(.*)/ { print $$2 }' $(MAKEFILE_LIST)
