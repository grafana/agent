#!/usr/bin/env bash
MANIFEST_BRANCH=v0.23.0
MANIFEST_URL=${MANIFEST_URL:-https://raw.githubusercontent.com/grafana/agent/${MANIFEST_BRANCH}/production/kubernetes/agent-bare.yaml}
NAMESPACE=${NAMESPACE:-default}

export NAMESPACE

curl -fsSL "$MANIFEST_URL" | envsubst
