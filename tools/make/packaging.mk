# Makefile.packaging adds release-packaging-specific targets.

PARENT_MAKEFILE := $(firstword $(MAKEFILE_LIST))

dist: dist-agent dist-agentctl dist-packages

dist-agent: dist/agent-linux-amd64       \
            dist/agent-linux-arm64       \
            dist/agent-linux-armv6       \
            dist/agent-linux-armv7       \
            dist/agent-linux-ppc64le     \
            dist/agent-darwin-amd64      \
            dist/agent-darwin-arm64      \
            dist/agent-windows-amd64.exe \
            dist/agent-windows-installer.exe

# Ensure there's no ebpf support in the linux/amd64 builds.
dist/agent-linux-amd64: GO_TAGS      += noebpf
dist/agent-linux-amd64: GOOS         := linux
dist/agent-linux-amd64: GOARCH       := amd64
dist/agent-linux-amd64: AGENT_BINARY := dist/agent-linux-amd64
dist/agent-linux-amd64:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-arm64: GOOS         := linux
dist/agent-linux-arm64: GOARCH       := arm64
dist/agent-linux-arm64: AGENT_BINARY := dist/agent-linux-arm64
dist/agent-linux-arm64:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-armv6: GOOS         := linux
dist/agent-linux-armv6: GOARCH       := arm
dist/agent-linux-armv6: GOARM        := 6
dist/agent-linux-armv6: AGENT_BINARY := dist/agent-linux-armv6
dist/agent-linux-armv6:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-armv7: GOOS         := linux
dist/agent-linux-armv7: GOARCH       := arm
dist/agent-linux-armv7: GOARM        := 7
dist/agent-linux-armv7: AGENT_BINARY := dist/agent-linux-armv7
dist/agent-linux-armv7:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-ppc64le: GOOS         := linux
dist/agent-linux-ppc64le: GOARCH       := ppc64le
dist/agent-linux-ppc64le: AGENT_BINARY := dist/agent-linux-ppc64le
dist/agent-linux-ppc64le:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-darwin-amd64: GOOS         := darwin
dist/agent-darwin-amd64: GOARCH       := amd64
dist/agent-darwin-amd64: AGENT_BINARY := dist/agent-darwin-amd64
dist/agent-darwin-amd64:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-darwin-arm64: GOOS         := darwin
dist/agent-darwin-arm64: GOARCH       := arm64
dist/agent-darwin-arm64: AGENT_BINARY := dist/agent-darwin-arm64
dist/agent-darwin-arm64:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-windows-amd64.exe: GOOS         := windows
dist/agent-windows-amd64.exe: GOARCH       := amd64
dist/agent-windows-amd64.exe: AGENT_BINARY := dist/agent-windows-amd64.exe
dist/agent-windows-amd64.exe:
	GO_TAGS=$(GO_TAGS) GOOS=$(GOOS) GOARCH=$(GOARCH) AGENT_BINARY=$(AGENT_BINARY) $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-windows-installer.exe: dist/agent-windows-amd64.exe
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	makensis -V4 -DVERSION=$(VERSION) -DOUT="../../dist/grafana-agent-installer.exe" ./packaging/windows/install_script.nsis
endif
