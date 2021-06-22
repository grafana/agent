#!/usr/bin/env bash
# shellcheck shell=bash

#
# install-bare.sh is a installer for the Agent without the ConfigMap. It is
# useful for being integrated into an installation process that provides a
# ConfigMap following the installation of Kubernetes components.
#

check_installed() {
  if ! type "$1" >/dev/null 2>&1; then
    echo "error: $1 not installed" >&2
    exit 1
  fi
}

check_installed curl
check_installed envsubst

MANIFEST_BRANCH=v0.16.1
MANIFEST_URL=${MANIFEST_URL:-https://raw.githubusercontent.com/grafana/agent/${MANIFEST_BRANCH}/production/kubernetes/agent-bare.yaml}
NAMESPACE=${NAMESPACE:-default}

export NAMESPACE

curl -fsSL "$MANIFEST_URL" | envsubst
