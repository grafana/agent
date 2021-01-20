#!/usr/bin/env bash

#
# install.sh is a really basic installer for the agent. It uses the existing
# Kubernetes YAML example and envsubst to fill in the details for remote write
# URL, username, and password.
#
# There are three ways to provide the inputs for installation:
#
# 1. Environment variables (REMOTE_WRITE_URL, REGION, ROLE_ARN, NAMESPACE)
#
# 2. Flags (-l for remote write URL, -n for namespace, -a for ARN, -r for region)
#
# 3. stdin from prompts
#
# Flags override environment variables, and stdin is used as a fallback if a
# value wasn't given from a flag or environment variable. An empty username
# or password is acceptable, as it disables basic auth. However, a remote write
# URL must always be provided.
#

check_installed() {
  if ! type $1 >/dev/null 2>&1; then
    echo "error: $1 not installed" >&2
    exit 1
  fi
}

check_installed curl
check_installed envsubst

MANIFEST_BRANCH=v0.11.0
MANIFEST_URL=${MANIFEST_URL:-https://raw.githubusercontent.com/grafana/agent/${MANIFEST_BRANCH}/production/kubernetes/agent-sigv4.yaml}
NAMESPACE=${NAMESPACE:-default}
ROLE_ARN=${ROLE_ARN:-}

while getopts "l:u:p:" opt; do
  case "$opt" in
    l)
      REMOTE_WRITE_URL=$OPTARG
      ;;
    n)
      NAMESPACE=$OPTARG
      ;;
    a)
      ROLE_ARN=$OPTARG
      ;;
    r)
      REGION=$OPTARG
      ;;
    ?)
      echo "usage: $(basename $0) [-l remote write url] [-n namespace]" >&2
      exit 1
      ;;
  esac
done

if [ -z "${REMOTE_WRITE_URL}" ]; then
  read -sp 'Enter your remote write URL: ' REMOTE_WRITE_URL
  printf $'\n' >&2

  if [ -z "${REMOTE_WRITE_URL}" ]; then
    echo "error: REMOTE_WRITE_URL must be provided by flag, env, or stdin" >&2
    exit 1
  fi
fi

if [ -z "${REGION}" ]; then
  read -sp 'Enter the remote write region: ' REGION
  printf $'\n' >&2

  if [ -z "${REGION}" ]; then
    echo "error: REGION must be provided by flag, env, or stdin" >&2
    exit 1
  fi
fi


export NAMESPACE
export REMOTE_WRITE_URL
export ROLE_ARN
export REGION

curl -fsSL $MANIFEST_URL | envsubst
