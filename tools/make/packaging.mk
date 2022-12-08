# Makefile.packaging adds release-packaging-specific targets.

PARENT_MAKEFILE := $(firstword $(MAKEFILE_LIST))

.PHONY: dist clean-dist
dist: dist-agent-binaries dist-agentctl-binaries dist-packages dist-agent-installer

clean-dist:
	rm -rf dist

# Used for passing through environment variables to sub-makes.
#
# NOTE(rfratto): This *must* use `=` instead of `:=` so it's expanded at
# reference time. Earlier iterations of this file had each target explicitly
# list these, but it's too easy to forget to set on so this is used to ensure
# everything needed is always passed through.
PACKAGING_VARS = RELEASE_BUILD=1 GO_TAGS="$(GO_TAGS)" GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM)

#
# agent release binaries
#

dist-agent-binaries: dist/agent-linux-amd64   \
                     dist/agent-linux-arm64   \
                     dist/agent-linux-armv6   \
                     dist/agent-linux-armv7   \
                     dist/agent-linux-ppc64le \
                     dist/agent-darwin-amd64  \
                     dist/agent-darwin-arm64  \
                     dist/agent-windows-amd64.exe \
                     dist/agent-freebsd-amd64

dist/agent-linux-amd64: GO_TAGS += builtinassets
dist/agent-linux-amd64: GOOS    := linux
dist/agent-linux-amd64: GOARCH  := amd64
dist/agent-linux-amd64: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-arm64: GO_TAGS += builtinassets
dist/agent-linux-arm64: GOOS    := linux
dist/agent-linux-arm64: GOARCH  := arm64
dist/agent-linux-arm64: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-armv6: GO_TAGS += builtinassets
dist/agent-linux-armv6: GOOS    := linux
dist/agent-linux-armv6: GOARCH  := arm
dist/agent-linux-armv6: GOARM   := 6
dist/agent-linux-armv6: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-armv7: GO_TAGS += builtinassets
dist/agent-linux-armv7: GOOS    := linux
dist/agent-linux-armv7: GOARCH  := arm
dist/agent-linux-armv7: GOARM   := 7
dist/agent-linux-armv7: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-linux-ppc64le: GO_TAGS += builtinassets
dist/agent-linux-ppc64le: GOOS    := linux
dist/agent-linux-ppc64le: GOARCH  := ppc64le
dist/agent-linux-ppc64le: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-darwin-amd64: GO_TAGS += builtinassets
dist/agent-darwin-amd64: GOOS    := darwin
dist/agent-darwin-amd64: GOARCH  := amd64
dist/agent-darwin-amd64: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-darwin-arm64: GO_TAGS += builtinassets
dist/agent-darwin-arm64: GOOS    := darwin
dist/agent-darwin-arm64: GOARCH  := arm64
dist/agent-darwin-arm64: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-windows-amd64.exe: GO_TAGS += builtinassets
dist/agent-windows-amd64.exe: GOOS    := windows
dist/agent-windows-amd64.exe: GOARCH  := amd64
dist/agent-windows-amd64.exe: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

dist/agent-freebsd-amd64: GO_TAGS += builtinassets
dist/agent-freebsd-amd64: GOOS    := freebsd
dist/agent-freebsd-amd64: GOARCH  := amd64
dist/agent-freebsd-amd64: generate-ui
	$(PACKAGING_VARS) AGENT_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agent

#
# agentctl release binaries.
#

dist-agentctl-binaries: dist/agentctl-linux-amd64   \
                        dist/agentctl-linux-arm64   \
                        dist/agentctl-linux-armv6   \
                        dist/agentctl-linux-armv7   \
                        dist/agentctl-linux-ppc64le \
                        dist/agentctl-darwin-amd64  \
                        dist/agentctl-darwin-arm64  \
                        dist/agentctl-windows-amd64.exe \
                        dist/agentctl-freebsd-amd64

dist/agentctl-linux-amd64: GOOS    := linux
dist/agentctl-linux-amd64: GOARCH  := amd64
dist/agentctl-linux-amd64:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-linux-arm64: GOOS   := linux
dist/agentctl-linux-arm64: GOARCH := arm64
dist/agentctl-linux-arm64:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-linux-armv6: GOOS   := linux
dist/agentctl-linux-armv6: GOARCH := arm
dist/agentctl-linux-armv6: GOARM  := 6
dist/agentctl-linux-armv6:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-linux-armv7: GOOS   := linux
dist/agentctl-linux-armv7: GOARCH := arm
dist/agentctl-linux-armv7: GOARM  := 7
dist/agentctl-linux-armv7:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-linux-ppc64le: GOOS   := linux
dist/agentctl-linux-ppc64le: GOARCH := ppc64le
dist/agentctl-linux-ppc64le:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-darwin-amd64: GOOS   := darwin
dist/agentctl-darwin-amd64: GOARCH := amd64
dist/agentctl-darwin-amd64:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-darwin-arm64: GOOS   := darwin
dist/agentctl-darwin-arm64: GOARCH := arm64
dist/agentctl-darwin-arm64:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-windows-amd64.exe: GOOS   := windows
dist/agentctl-windows-amd64.exe: GOARCH := amd64
dist/agentctl-windows-amd64.exe:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

dist/agentctl-freebsd-amd64: GO_TAGS += builtinassets
dist/agentctl-freebsd-amd64: GOOS    := freebsd
dist/agentctl-freebsd-amd64: GOARCH  := amd64
dist/agentctl-freebsd-amd64:
	$(PACKAGING_VARS) AGENTCTL_BINARY=$@ $(MAKE) -f $(PARENT_MAKEFILE) agentctl

#
# DEB and RPM packages.
#

ENVIRONMENT_FILE_rpm := /etc/sysconfig/grafana-agent
ENVIRONMENT_FILE_deb := /etc/default/grafana-agent

# generate_fpm(deb|rpm, package arch, agent arch, output file)
define generate_fpm =
	fpm -s dir -v $(PACKAGE_VERSION) -a $(2) \
		-n grafana-agent --iteration $(PACKAGE_RELEASE) -f \
		--log error \
		--license "Apache 2.0" \
		--vendor "Grafana Labs" \
		--url "https://github.com/grafana/agent" \
		-t $(1) \
		--after-install packaging/$(1)/control/postinst \
		--before-remove packaging/$(1)/control/prerm \
		--config-files /etc/grafana-agent.yaml \
		--config-files $(ENVIRONMENT_FILE_$(1)) \
		--package $(4) \
			dist/agent-linux-$(3)=/usr/bin/grafana-agent \
			dist/agentctl-linux-$(3)=/usr/bin/grafana-agentctl \
			packaging/grafana-agent.yaml=/etc/grafana-agent.yaml \
			packaging/environment-file=$(ENVIRONMENT_FILE_$(1)) \
			packaging/$(1)/grafana-agent.service=/usr/lib/systemd/system/grafana-agent.service
endef

PACKAGE_VERSION := $(patsubst v%,%,$(VERSION))
PACKAGE_RELEASE := 1
PACKAGE_PREFIX  := dist/grafana-agent-$(PACKAGE_VERSION)-$(PACKAGE_RELEASE)

.PHONY: dist-packages
dist-packages: dist-packages-amd64 \
               dist-packages-arm64 \
               dist-packages-armv6 \
               dist-packages-armv7 \
               dist-packages-arm64 \
               dist-packages-ppc64le

.PHONY: dist-packages-amd64
dist-packages-amd64: dist/agent-linux-amd64 dist/agentctl-linux-amd64
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(call generate_fpm,deb,amd64,amd64,$(PACKAGE_PREFIX).amd64.deb)
	$(call generate_fpm,rpm,x86_64,amd64,$(PACKAGE_PREFIX).amd64.rpm)
endif

.PHONY: dist-packages-arm64
dist-packages-arm64: dist/agent-linux-arm64 dist/agentctl-linux-arm64
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(call generate_fpm,deb,arm64,arm64,$(PACKAGE_PREFIX).arm64.deb)
	$(call generate_fpm,rpm,aarch64,arm64,$(PACKAGE_PREFIX).arm64.rpm)
endif

# There's no RPM for armv6 so only debs are produced.
.PHONY: dist-packages-armv6
dist-packages-armv6: dist/agent-linux-armv6 dist/agentctl-linux-armv6
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(call generate_fpm,deb,armhf,armv6,$(PACKAGE_PREFIX).armv6.deb)
endif

.PHONY: dist-packages-armv7
dist-packages-armv7: dist/agent-linux-armv7 dist/agentctl-linux-armv7
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(call generate_fpm,deb,armhf,armv7,$(PACKAGE_PREFIX).armv7.deb)
	$(call generate_fpm,rpm,armhfp,armv7,$(PACKAGE_PREFIX).armv7.rpm)
endif

.PHONY: dist-packages-ppc64le
dist-packages-ppc64le: dist/agent-linux-ppc64le dist/agentctl-linux-ppc64le
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	$(call generate_fpm,deb,ppc64el,ppc64le,$(PACKAGE_PREFIX).ppc64el.deb)
	$(call generate_fpm,rpm,ppc64le,ppc64le,$(PACKAGE_PREFIX).ppc64le.rpm)
endif

#
# Windows installer
#

# TODO(rfratto): update the install_script.nsis so we don't need to copy assets
# over into the packaging/windows folder.
.PHONY: dist-agent-installer
dist-agent-installer: dist/agent-windows-amd64.exe
ifeq ($(USE_CONTAINER),1)
	$(RERUN_IN_CONTAINER)
else
	cp ./dist/agent-windows-amd64.exe ./packaging/windows
	cp LICENSE ./packaging/windows
	makensis -V4 -DVERSION=$(VERSION) -DOUT="../../dist/grafana-agent-installer.exe" ./packaging/windows/install_script.nsis
endif
