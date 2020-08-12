# TODO(rfratto): docker images

.DEFAULT_GOAL := all
.PHONY: all agent agentctl check-mod int test clean cmd/agent/agent cmd/agentctl/agentctl protos

SHELL = /usr/bin/env bash

#############
# Variables #
#############

# When the value of empty, no -mod parameter will be passed to go.
# For Go 1.13, "readonly" and "vendor" can be used here.
# In Go 1.14, "vendor" and "mod" can be used instead.
GOMOD?=vendor
ifeq ($(strip $(GOMOD)),) # Is empty?
	MOD_FLAG=
	GOLANGCI_ARG=
else
	MOD_FLAG=-mod=$(GOMOD)
	GOLANGCI_ARG=--modules-download-mode=$(GOMOD)
endif

# Certain aspects of the build are done in containers for consistency.
# If you have the correct tools installed and want to speed up development,
# run make BUILD_IN_CONTAINER=false <target>, or you can set BUILD_IN_CONTAINER=true
# as an environment variable.
BUILD_IN_CONTAINER ?= true
BUILD_IMAGE_VERSION := 0.9.0

# Enables the binary to be built with optimizations (i.e., doesn't strip the image of
# symbols, etc.)
RELEASE_BUILD ?= false

# Docker image info
IMAGE_PREFIX ?= grafana
IMAGE_TAG ?= $(shell ./tools/image-tag)

# Version info for binaries
GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

# When running find there's a set of directories we'll never care about; we
# define the list here to make scanning faster.
DONT_FIND := -name tools -prune -o -name vendor -prune -o -name .git -prune -o -name .cache -prune -o -name .pkg -prune -o

# Build flags
VPREFIX        := github.com/grafana/agent/pkg/build
GO_LDFLAGS     := -X $(VPREFIX).Branch=$(GIT_BRANCH) -X $(VPREFIX).Version=$(IMAGE_TAG) -X $(VPREFIX).Revision=$(GIT_REVISION) -X $(VPREFIX).BuildUser=$(shell whoami)@$(shell hostname) -X $(VPREFIX).BuildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_FLAGS       := -ldflags "-extldflags \"-static\" -s -w $(GO_LDFLAGS)" -tags "netgo static_build" $(MOD_FLAG)
DEBUG_GO_FLAGS := -gcflags "all=-N -l" -ldflags "-extldflags \"-static\" $(GO_LDFLAGS)" -tags "netgo static_build" $(MOD_FLAG)

# We need a separate set of flags for CGO, where building with -static can
# cause problems with some C libraries.
CGO_FLAGS := -ldflags "-s -w $(GO_LDFLAGS)" -tags "netgo" $(MOD_FLAG)
DEBUG_CGO_FLAGS := -gcflags "all=-N -l" -ldflags "-s -w $(GO_LDFLAGS)" -tags "netgo" $(MOD_FLAG)

# If we're not building the release, use the debug flags instead.
ifeq ($(RELEASE_BUILD),false)
GO_FLAGS = $(DEBUG_GO_FLAGS)
endif

NETGO_CHECK = @strings $@ | grep cgo_stub\\\.go >/dev/null || { \
       rm $@; \
       echo "\nYour go standard library was built without the 'netgo' build tag."; \
       echo "To fix that, run"; \
       echo "    sudo go clean -i net"; \
       echo "    sudo go install -tags netgo std"; \
       false; \
}

# Protobuf files
PROTO_DEFS := $(shell find . $(DONT_FIND) -type f -name '*.proto' -print)
PROTO_GOS := $(patsubst %.proto,%.pb.go,$(PROTO_DEFS))

#############
# Protobufs #
#############

protos: $(PROTO_GOS)

# Use with care; this signals to make that the proto definitions don't need recompiling.
touch-protos:
	for proto in $(PROTO_GOS); do [ -f "./$${proto}" ] && touch "$${proto}" && echo "touched $${proto}"; done

%.pb.go: $(PROTO_DEFS)
# We use loki-build-image here which expects /src/loki so we bind mount the agent
# repo to /src/loki just for building the protobufs.
ifeq ($(BUILD_IN_CONTAINER),true)
	@mkdir -p $(shell pwd)/.pkg
	@mkdir -p $(shell pwd)/.cache
	docker run -i \
		-v $(shell pwd)/.cache:/go/cache \
		-v $(shell pwd)/.pkg:/go/pkg \
		-v $(shell pwd):/src/loki \
		$(IMAGE_PREFIX)/loki-build-image:$(BUILD_IMAGE_VERSION) $@;
else
	protoc -I .:./vendor:./$(@D) --gogoslick_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,plugins=grpc,paths=source_relative:./ ./$(patsubst %.pb.go,%.proto,$@);
endif

###################
# Primary Targets #
###################
all: protos agent agentctl
agent: cmd/agent/agent
agentctl: cmd/agentctl/agentctl

cmd/agent/agent: cmd/agent/main.go
	CGO_ENABLED=1 go build $(CGO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

cmd/agentctl/agentctl: cmd/agentctl/main.go
	CGO_ENABLED=0 go build $(GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

agent-image:
	docker build --build-arg RELEASE_BUILD=$(RELEASE_BUILD)  --build-arg IMAGE_TAG=$(IMAGE_TAG) \
		-t $(IMAGE_PREFIX)/agent:latest -f cmd/agent/Dockerfile .
	docker tag $(IMAGE_PREFIX)/agent:latest $(IMAGE_PREFIX)/agent:$(IMAGE_TAG)

agentctl-image:
	docker build --build-arg RELEASE_BUILD=$(RELEASE_BUILD)  --build-arg IMAGE_TAG=$(IMAGE_TAG) \
		-t $(IMAGE_PREFIX)/agentctl:latest -f cmd/agentctl/Dockerfile .
	docker tag $(IMAGE_PREFIX)/agentctl:latest $(IMAGE_PREFIX)/agentctl:$(IMAGE_TAG)

push-agent-image:
	docker push $(IMAGE_PREFIX)/agent:latest
	docker push $(IMAGE_PREFIX)/agent:$(IMAGE_TAG)

push-agentctl-image:
	docker push $(IMAGE_PREFIX)/agentctl:latest
	docker push $(IMAGE_PREFIX)/agentctl:$(IMAGE_TAG)

install:
	CGO_ENABLED=1 go install $(CGO_FLAGS) ./cmd/agent
	CGO_ENABLED=0 go install $(GO_FLAGS) ./cmd/agentctl

#######################
# Development targets #
#######################

lint:
	GO111MODULE=on GOGC=10 golangci-lint run -v $(GOLANGCI_ARG)

# We have to run test twice: once for all packages with -race and then once more without -race
# for packages that have known race detection issues
test:
	GOGC=10 go test $(MOD_FLAG) -race -cover -coverprofile=cover.out -p=4 ./...
	GOGC=10 go test $(MOD_FLAG) -cover -coverprofile=cover-norace.out -p=4 ./pkg/integrations/node_exporter

clean:
	rm -rf cmd/agent/agent
	go clean $(MOD_FLAG) ./...

example-kubernetes:
	cd production/kubernetes/build && bash build.sh

example-dashboards:
	cd example/docker-compose/grafana/dashboards && \
		jsonnet template.jsonnet -J ../../vendor -m .

#############
# Releasing #
#############

seego = docker run --rm -t -v "$(CURDIR):$(CURDIR)" -w "$(CURDIR)" -e "CGO_ENABLED=$$CGO_ENABLED" -e "GOOS=$$GOOS" -e "GOARCH=$$GOARCH" -e "GOARM=$$GOARM" rfratto/seego

# dist builds the agent and agentctl for all different supported platforms.
# Most of these platforms need CGO_ENABLED=1, but to simplify things we'll
# use CGO_ENABLED for all of them. We define them all as separate targets
# to allow for parallelization with make -jX.
#
# We use rfratto/seego for building these cross-platform images. seego provides
# a docker image with gcc toolchains for all of these platforms.
dist: dist-agent dist-agentctl
	for i in dist/*; do zip -j -m $$i.zip $$i; done
	pushd dist && sha256sum * > SHA256SUMS && popd
.PHONY: dist

dist-agent: dist/agent-linux-amd64 dist/agent-darwin-amd64 dist/agent-freebsd-amd64 dist/agent-windows-amd64.exe
dist/agent-linux-amd64:
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agent
dist/agent-darwin-amd64:
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agent
dist/agent-freebsd-amd64:
	@CGO_ENABLED=1 GOOS=freebsd GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agent
dist/agent-windows-amd64.exe:
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agent

dist-agentctl: dist/agentctl-linux-amd64 dist/agentctl-darwin-amd64 dist/agentctl-freebsd-amd64 dist/agentctl-windows-amd64.exe
dist/agentctl-linux-amd64:
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agentctl
dist/agentctl-darwin-amd64:
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agentctl
dist/agentctl-freebsd-amd64:
	@CGO_ENABLED=1 GOOS=freebsd GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agentctl
dist/agentctl-windows-amd64.exe:
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64; $(seego) build $(CGO_FLAGS) -o $@ ./cmd/agentctl

clean-dist:
	rm -rf dist
.PHONY: clean

publish: dist
	./tools/release
