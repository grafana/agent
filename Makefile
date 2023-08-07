## Build, test, and generate code for various parts of Grafana Agent.
##
## At least Go 1.19, git, and a moderately recent version of Docker is required
## to be able to use the Makefile. This list isn't exhaustive and there are other
## dependencies for the generate-* targets. If you do not have the full list of
## build dependencies, you may set USE_CONTAINER=1 to proxy build commands to a
## build container.
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
##   binaries                Compiles all binaries.
##   agent                   Compiles cmd/grafana-agent to $(AGENT_BINARY)
##   agent-boringcrypto      Compiles cmd/grafana-agent with GOEXPERIMENT=boringcrypto to $(AGENT_BORINGCRYPTO_BINARY)
##   agent-flow              Compiles cmd/grafana-agent-flow to $(FLOW_BINARY)
##   agent-service           Compiles cmd/grafana-agent-service to $(SERVICE_BINARY)
##   agentctl                Compiles cmd/grafana-agentctl to $(AGENTCTL_BINARY)
##   operator                Compiles cmd/grafana-agent-operator to $(OPERATOR_BINARY)
##   crow                    Compiles tools/crow to $(CROW_BINARY)
##   smoke                   Compiles tools/smoke to $(SMOKE_BINARY)
##
## Targets for building Docker images:
##
##   images                   Builds all Docker images.
##   agent-image              Builds agent Docker image.
##   agent-boringcrypto-image Builds agent Docker image with boringcrypto.
##   agentctl-image           Builds agentctl Docker image.
##   operator-image           Builds operator Docker image.
##   crow-image               Builds crow Docker image.
##   smoke-image              Builds smoke test Docker image.
##
## Targets for packaging:
##
##   dist                   Produce release assets for everything.
##   dist-agent-binaries    Produce release-ready agent binaries.
##   dist-agentctl-binaries Produce release-ready agentctl binaries.
##   dist-packages          Produce release-ready DEB and RPM packages.
##   dist-agent-installer   Produce a Windows installer for Grafana Agent.
##
## Targets for generating assets:
##
##   generate             Generate everything.
##   generate-crds        Generate Grafana Agent Operator CRDs ands its documentation.
##   generate-drone       Generate the Drone YAML from Jsonnet.
##   generate-helm-docs   Generate Helm chart documentation.
##   generate-helm-tests  Generate Helm chart tests.
##   generate-manifests   Generate production/kubernetes YAML manifests.
##   generate-dashboards  Generate dashboards in example/docker-compose after
##                        changing Jsonnet.
##   generate-protos      Generate protobuf files.
##   generate-ui          Generate the UI assets.
##
## Other targets:
##
##   build-container-cache  Create a cache for the build container to speed up
##                          subsequent proxied builds
##   drone                  Sign Drone CI config (maintainers only)
##   clean                  Clean caches and built binaries
##   help                   Displays this message
##   info                   Print Makefile-specific environment variables
##
## Environment variables:
##
##   USE_CONTAINER              Set to 1 to enable proxying commands to build container
##   AGENT_IMAGE                Image name:tag built by `make agent-image`
##   AGENTCTL_IMAGE             Image name:tag built by `make agentctl-image`
##   OPERATOR_IMAGE             Image name:tag built by `make operator-image`
##   CROW_IMAGE                 Image name:tag built by `make crow-image`
##   SMOKE_IMAGE                Image name:tag built by `make smoke-image`
##   BUILD_IMAGE                Image name:tag used by USE_CONTAINER=1
##   AGENT_BINARY               Output path of `make agent` (default build/grafana-agent)
##   AGENT_BORINGCRYPTO_BINARY  Output path of `make agent-boringcrypto` (default build/grafana-agent-boringcrypto)
##   FLOW_BINARY                Output path of `make agent-flow` (default build/grafana-agent-flow)
##   SERVICE_BINARY             Output path of `make agent-service` (default build/grafana-agent-service)
##   AGENTCTL_BINARY            Output path of `make agentctl` (default build/grafana-agentctl)
##   OPERATOR_BINARY            Output path of `make operator` (default build/grafana-agent-operator)
##   CROW_BINARY                Output path of `make crow` (default build/grafana-agent-crow)
##   SMOKE_BINARY               Output path of `make smoke` (default build/grafana-agent-smoke)
##   GOOS                       Override OS to build binaries for
##   GOARCH                     Override target architecture to build binaries for
##   GOARM                      Override ARM version (6 or 7) when GOARCH=arm
##   CGO_ENABLED                Set to 0 to disable Cgo for binaries.
##   RELEASE_BUILD              Set to 1 to build release binaries.
##   VERSION                    Version to inject into built binaries.
##   GO_TAGS                    Extra tags to use when building.
##   DOCKER_PLATFORM            Overrides platform to build Docker images for (defaults to host platform).
##   GOEXPERIMENT               Used to enable features, most likely boringcrypto via GOEXPERIMENT=boringcrypto.

include tools/make/*.mk

AGENT_IMAGE                             ?= grafana/agent:latest
AGENT_BORINGCRYPTO_IMAGE                ?= grafana/agent-boringcrypto:latest
AGENTCTL_IMAGE                          ?= grafana/agentctl:latest
OPERATOR_IMAGE                          ?= grafana/agent-operator:latest
CROW_IMAGE                              ?= us.gcr.io/kubernetes-dev/grafana/agent-crow:latest
SMOKE_IMAGE                             ?= us.gcr.io/kubernetes-dev/grafana/agent-smoke:latest
AGENT_BINARY                            ?= build/grafana-agent
AGENT_BORINGCRYPTO_BINARY               ?= build/grafana-agent-boringcrypto
FLOW_BINARY                             ?= build/grafana-agent-flow
SERVICE_BINARY                          ?= build/grafana-agent-service
AGENTCTL_BINARY                         ?= build/grafana-agentctl
OPERATOR_BINARY                         ?= build/grafana-agent-operator
CROW_BINARY                             ?= build/agent-crow
SMOKE_BINARY                            ?= build/agent-smoke
AGENTLINT_BINARY                        ?= build/agentlint
GOOS                                    ?= $(shell go env GOOS)
GOARCH                                  ?= $(shell go env GOARCH)
GOARM                                   ?= $(shell go env GOARM)
CGO_ENABLED                             ?= 1
RELEASE_BUILD                           ?= 0
GOEXPERIMENT                            ?= $(shell go env GOEXPERIMENT)

# List of all environment variables which will propagate to the build
# container. USE_CONTAINER must _not_ be included to avoid infinite recursion.
PROPAGATE_VARS := \
    AGENT_IMAGE AGENTCTL_IMAGE OPERATOR_IMAGE CROW_IMAGE SMOKE_IMAGE \
    BUILD_IMAGE GOOS GOARCH GOARM CGO_ENABLED RELEASE_BUILD \
    AGENT_BINARY AGENT_BORINGCRYPTO_BINARY FLOW_BINARY AGENTCTL_BINARY OPERATOR_BINARY \
    CROW_BINARY SMOKE_BINARY VERSION GO_TAGS GOEXPERIMENT

#
# Constants for targets
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
DEBUG_GO_FLAGS   := -ldflags "$(GO_LDFLAGS)" -tags "$(GO_TAGS)"
RELEASE_GO_FLAGS := -ldflags "-s -w $(GO_LDFLAGS)" -tags "$(GO_TAGS)"

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
lint: agentlint
	golangci-lint run -v --timeout=10m
	$(AGENTLINT_BINARY) ./...

.PHONY: test
# We have to run test twice: once for all packages with -race and then once
# more without -race for packages that have known race detection issues.
test:
	$(GO_ENV) go test $(GO_FLAGS) -race ./...
	$(GO_ENV) go test $(GO_FLAGS) ./pkg/integrations/node_exporter ./pkg/logs ./pkg/operator ./pkg/util/k8s ./component/otelcol/processor/tail_sampling ./component/loki/source/file

test-packages:
	docker pull $(BUILD_IMAGE)
	go test -tags=packaging  ./packaging

#
# Targets for building binaries
#

.PHONY: binaries agent agent-boringcrypto agent-flow agentctl operator crow smoke
binaries: agent agent-boringcrypto agent-flow agentctl operator crow smoke

agent:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(AGENT_BINARY) ./cmd/grafana-agent
endif

agent-boringcrypto:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	GOEXPERIMENT=boringcrypto $(GO_ENV) go build $(GO_FLAGS) -o $(AGENT_BORINGCRYPTO_BINARY) ./cmd/grafana-agent
endif


agent-flow:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(FLOW_BINARY) ./cmd/grafana-agent-flow
endif

# agent-service is not included in binaries since it's Windows-only.
agent-service:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(SERVICE_BINARY) ./cmd/grafana-agent-service
endif

agentctl:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(AGENTCTL_BINARY) ./cmd/grafana-agentctl
endif

operator:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(GO_ENV) go build $(GO_FLAGS) -o $(OPERATOR_BINARY) ./cmd/grafana-agent-operator
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

agentlint:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	cd ./tools/agentlint && $(GO_ENV) go build $(GO_FLAGS) -o ../../$(AGENTLINT_BINARY) .
endif

#
# Targets for building Docker images
#

DOCKER_FLAGS := --build-arg RELEASE_BUILD=$(RELEASE_BUILD) --build-arg VERSION=$(VERSION)

ifneq ($(DOCKER_PLATFORM),)
DOCKER_FLAGS += --platform=$(DOCKER_PLATFORM)
endif

.PHONY: images agent-image agentctl-image operator-image crow-image smoke-image
images: agent-image agentctl-image operator-image crow-image smoke-image

agent-image:
	DOCKER_BUILDKIT=1 docker build $(DOCKER_FLAGS) -t $(AGENT_IMAGE) -f cmd/grafana-agent/Dockerfile .
agentctl-image:
	DOCKER_BUILDKIT=1 docker build $(DOCKER_FLAGS) -t $(AGENTCTL_IMAGE) -f cmd/grafana-agentctl/Dockerfile .
agent-boringcrypto-image:
	DOCKER_BUILDKIT=1 docker build $(DOCKER_FLAGS) --build-arg GOEXPERIMENT=boringcrypto -t $(AGENT_BORINGCRYPTO_IMAGE) -f cmd/grafana-agent/Dockerfile .
operator-image:
	DOCKER_BUILDKIT=1 docker build $(DOCKER_FLAGS) -t $(OPERATOR_IMAGE) -f cmd/grafana-agent-operator/Dockerfile .
crow-image:
	DOCKER_BUILDKIT=1 docker build $(DOCKER_FLAGS) -t $(CROW_IMAGE) -f tools/crow/Dockerfile .
smoke-image:
	DOCKER_BUILDKIT=1 docker build $(DOCKER_FLAGS) -t $(SMOKE_IMAGE) -f tools/smoke/Dockerfile .

#
# Targets for generating assets
#

.PHONY: generate generate-crds generate-drone generate-helm-docs generate-helm-tests generate-manifests generate-dashboards generate-protos generate-ui
generate: generate-crds generate-drone generate-helm-docs generate-helm-tests generate-manifests generate-dashboards generate-protos generate-ui

generate-crds:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	bash ./tools/generate-crds.bash
	gen-crd-api-reference-docs -config tools/gen-crd-docs/config.json -api-dir "github.com/grafana/agent/pkg/operator/apis/monitoring/" -out-file docs/sources/operator/api.md -template-dir tools/gen-crd-docs/template
endif

generate-drone:
	drone jsonnet -V BUILD_IMAGE_VERSION=$(BUILD_IMAGE_VERSION) --stream --format --source .drone/drone.jsonnet --target .drone/drone.yml

generate-helm-docs:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	cd operations/helm/charts/grafana-agent && helm-docs
endif

generate-helm-tests:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	bash ./operations/helm/scripts/rebuild-tests.sh
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

generate-ui:
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	cd ./web/ui && yarn --network-timeout=1200000 && yarn run build
endif

#
# Other targets
#
# build-container-cache and clean-build-container-cache are defined in
# Makefile.build-container.

# Drone signs the yaml, you will need to specify DRONE_TOKEN, which can be
# found by logging into your profile in Drone.
#
# This will only work for maintainers.
.PHONY: drone
drone: generate-drone
	drone lint .drone/drone.yml --trusted
	drone --server https://drone.grafana.net sign --save grafana/agent .drone/drone.yml

.PHONY: clean
clean: clean-dist clean-build-container-cache
	rm -rf ./build/*

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
	@printf "GOEXPERIMENT    = $(GOEXPERIMENT)\n"

# awk magic to print out the comment block at the top of this file.
.PHONY: help
help:
	@awk 'BEGIN {FS="## "} /^##\s*(.*)/ { print $$2 }' $(MAKEFILE_LIST)
