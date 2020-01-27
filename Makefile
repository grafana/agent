# TODO(rfratto): docker images

.DEFAULT_GOAL := all
.PHONY: all agent check-mod int test clean

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

# Docker image info
IMAGE_PREFIX ?= grafana

# Version info for binaries
GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

# Build flags
VPREFIX        := github.com/grafana/agent/cmd/agent/build
GO_LDFLAGS     := -X $(VPREFIX).Branch=$(GIT_BRANCH) -X $(VPREFIX).Version=$(IMAGE_TAG) -X $(VPREFIX).Revision=$(GIT_REVISION) -X $(VPREFIX).BuildUser=$(shell whoami)@$(shell hostname) -X $(VPREFIX).BuildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_FLAGS       := -ldflags "-extldflags \"-static\" -s -w $(GO_LDFLAGS)" -tags netgo $(MOD_FLAG)
DEBUG_GO_FLAGS := -gcflags "all=-N -l" -ldflags "-extldflags \"-static\" $(GO_LDFLAGS)" -tags netgo $(MOD_FLAG)

NETGO_CHECK = @strings $@ | grep cgo_stub\\\.go >/dev/null || { \
       rm $@; \
       echo "\nYour go standard library was built without the 'netgo' build tag."; \
       echo "To fix that, run"; \
       echo "    sudo go clean -i net"; \
       echo "    sudo go install -tags netgo std"; \
       false; \
}

###################
# Primary Targets #
###################
all: agent
agent: cmd/agent/agent

cmd/agent/agent: cmd/agent/main.go
	CGO_ENABLED=0 go build $(GO_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

#######################
# Development targets #
#######################

lint:
	GO111MODULE=on GOGC=10 golangci-lint run -v $(GOLANGCI_ARG)

test: all
	GOGC=10 go test $(MOD_FLAG) -p=4 ./...

clean:
	rm -rf cmd/agent/agent
	go clean $(MOD_FLAG) ./...
