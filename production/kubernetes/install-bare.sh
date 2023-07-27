#!/usr/bin/env bash
# shellcheck shell=bash

#
# install-bare.sh is an installer for the Agent without a ConfigMap. It is
# used during the Grafana Cloud integrations wizard and is not recommended
# to be used directly. Instead of calling this script directly, please
# make a copy of ./agent-bare.yaml and modify it for your needs.
#
# Note that agent-bare.yaml does not have a ConfigMap, so the Grafana Agent
# will not launch until one is created. For more information on setting up
# a ConfigMap, please refer to:
#
# Metrics quickstart: https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_metrics/
# Logs quickstart: https://grafana.com/docs/grafana-cloud/quickstart/agent-k8s/k8s_agent_logs/
#

check_installed() {
  if ! type "$1" >/dev/null 2>&1; then
    echo "error: $1 not installed" >&2
    exit 1
  fi
}

check_installed curl
check_installed envsubst

MANIFEST_BRANCH=v0.35.2
MANIFEST_URL=${MANIFEST_URL:-https://raw.githubusercontent.com/grafana/agent/${MANIFEST_BRANCH}/production/kubernetes/agent-bare.yaml}
NAMESPACE=${NAMESPACE:-default}

export NAMESPACE

curl -fsSL "$MANIFEST_URL" | envsubst
