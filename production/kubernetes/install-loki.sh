#!/usr/bin/env bash
# shellcheck shell=bash

#
# install-loki.sh is a really basic installer for the agent. It uses the existing
# Kubernetes YAML example and envsubst to fill in the details for Loki write
# URL, username, and password.
#
# There are three ways to provide the inputs for installation:
#
# 1. Environment variables (LOKI_HOSTNAME, LOKI_USERNAME, LOKI_PASSWORD,
#    NAMESPACE)
#
# 2. Flags (-h for hostname, -u for username, -p for password, -n namespace)
#
# 3. stdin from prompts
#
# Flags override environment variables, and stdin is used as a fallback if a
# value wasn't given from a flag or environment variable. An empty username
# or password is acceptable, as it disables basic auth. However, a hostname must
# always be provided.
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
MANIFEST_URL=${MANIFEST_URL:-https://raw.githubusercontent.com/grafana/agent/${MANIFEST_BRANCH}/production/kubernetes/agent-loki.yaml}
NAMESPACE=${NAMESPACE:-default}

LOKI_USERNAME_SET=0
LOKI_PASSWORD_SET=0

while getopts "h:u:p:n:" opt; do
  case "$opt" in
    h)
      LOKI_HOSTNAME=$OPTARG
      ;;
    u)
      LOKI_USERNAME=$OPTARG
      LOKI_USERNAME_SET=1
      ;;
    p)
      LOKI_PASSWORD=$OPTARG
      LOKI_PASSWORD_SET=1
      ;;
    n)
      NAMESPACE=$OPTARG
      ;;
    ?)
      echo "usage: $(basename "$0") [-h Loki hostname] [-u Loki username] [-p Loki password] [-n namespace]" >&2
      exit 1
      ;;
  esac
done

if [ -z "${LOKI_HOSTNAME}" ]; then
  read -rsp 'Enter your Loki hostname (i.e., logs-us-central1.grafana.net): ' LOKI_HOSTNAME
  printf $'\n' >&2

  # We require a hostname for the agent; we don't do this same check for
  # the username and password as the remote Loki system may not have basic
  # auth enabled.
  if [ -z "${LOKI_HOSTNAME}" ]; then
    echo "error: LOKI_HOSTNAME must be provided by flag, env, or stdin" >&2
    exit 1
  fi
fi

if [ -z "${LOKI_USERNAME}" ] && [ "${LOKI_USERNAME_SET}" -eq 0 ]; then
  read -rsp 'Enter your Loki username: ' LOKI_USERNAME
  printf $'\n' >&2
fi

if [ -z "${LOKI_PASSWORD}" ] && [ "${LOKI_PASSWORD_SET}" -eq 0 ]; then
  read -rsp 'Enter your Loki password: ' LOKI_PASSWORD
  printf $'\n' >&2
fi

export NAMESPACE
export LOKI_HOSTNAME
export LOKI_USERNAME
export LOKI_PASSWORD

curl -fsSL "$MANIFEST_URL" | envsubst
