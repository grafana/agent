#!/usr/bin/env bash
# shellcheck shell=bash

#
# install.sh is a really basic installer for the agent. It uses the existing
# Kubernetes YAML example and envsubst to fill in the details for remote write
# URL, username, and password.
#
# There are three ways to provide the inputs for installation:
#
# 1. Environment variables (NAMESPACE, REMOTE_WRITE_URL, REMOTE_WRITE_USERNAME,
#    REMOTE_WRITE_PASSWORD)
#
# 2. Flags (-l for remote write URL, -u for remote write username, -p for remote
#    write password, -n for namespace)
#
# 3. stdin from prompts
#
# Flags override environment variables, and stdin is used as a fallback if a
# value wasn't given from a flag or environment variable. An empty username
# or password is acceptable, as it disables basic auth. However, a remote write
# URL must always be provided.
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
MANIFEST_URL=${MANIFEST_URL:-https://raw.githubusercontent.com/grafana/agent/${MANIFEST_BRANCH}/production/kubernetes/agent.yaml}
NAMESPACE=${NAMESPACE:-default}

REMOTE_WRITE_USERNAME_SET=0
REMOTE_WRITE_PASSWORD_SET=0

while getopts "l:u:p:n:" opt; do
  case "$opt" in
    l)
      REMOTE_WRITE_URL=$OPTARG
      ;;
    u)
      REMOTE_WRITE_USERNAME=$OPTARG
      REMOTE_WRITE_USERNAME_SET=1
      ;;
    p)
      REMOTE_WRITE_PASSWORD=$OPTARG
      REMOTE_WRITE_PASSWORD_SET=1
      ;;
    n)
      NAMESPACE=$OPTARG
      ;;
    ?)
      echo "usage: $(basename "$0") [-l remote write url] [-u remote write username] [-p remote write password] [-n namespace]" >&2
      exit 1
      ;;
  esac
done

if [ -z "${REMOTE_WRITE_URL}" ]; then
  read -rsp 'Enter your remote write URL: ' REMOTE_WRITE_URL
  printf $'\n' >&2

  # We require a remote write URL for the agent; we don't do this same check for
  # the username and password as the remote write system may not have basic
  # auth enabled.
  if [ -z "${REMOTE_WRITE_URL}" ]; then
    echo "error: REMOTE_WRITE_URL must be provided by flag, env, or stdin" >&2
    exit 1
  fi
fi

if [ -z "${REMOTE_WRITE_USERNAME}" ] && [ "${REMOTE_WRITE_USERNAME_SET}" -eq 0 ]; then
  read -rsp 'Enter your remote write username: ' REMOTE_WRITE_USERNAME
  printf $'\n' >&2
fi

if [ -z "${REMOTE_WRITE_PASSWORD}" ] && [ "${REMOTE_WRITE_PASSWORD_SET}" -eq 0 ]; then
  read -rsp 'Enter your remote write password: ' REMOTE_WRITE_PASSWORD
  printf $'\n' >&2
fi

export NAMESPACE
export REMOTE_WRITE_URL
export REMOTE_WRITE_USERNAME
export REMOTE_WRITE_PASSWORD

curl -fsSL "$MANIFEST_URL" | envsubst
