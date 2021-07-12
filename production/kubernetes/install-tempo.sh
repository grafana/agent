#!/usr/bin/env bash
# shellcheck shell=bash

#
# install-tempo.sh is a really basic installer for the agent. It uses the existing
# Kubernetes YAML example and envsubst to fill in the details for Tempo push
# URL, username, and password.
#
# There are three ways to provide the inputs for installation:
#
# 1. Environment variables (TEMPO_ENDPOINT, TEMPO_USERNAME, TEMPO_PASSWORD,
#    NAMESPACE)
#
# 2. Flags (-e for endpoint, -u for username, -p for password, -n for namespace)
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
MANIFEST_URL=${MANIFEST_URL:-https://raw.githubusercontent.com/grafana/agent/${MANIFEST_BRANCH}/production/kubernetes/agent-tempo.yaml}
NAMESPACE=${NAMESPACE:-default}

TEMPO_USERNAME_SET=0
TEMPO_PASSWORD_SET=0

while getopts "e:u:p:n:" opt; do
  case "$opt" in
    e)
      TEMPO_ENDPOINT=$OPTARG
      ;;
    u)
      TEMPO_USERNAME=$OPTARG
      TEMPO_USERNAME_SET=1
      ;;
    p)
      TEMPO_PASSWORD=$OPTARG
      TEMPO_PASSWORD_SET=1
      ;;
    n)
      NAMESPACE=$OPTARG
      ;;
    ?)
      echo "usage: $(basename "$0") [-e Tempo endpoint] [-u Tempo username] [-p Tempo password] [-n namespace]" >&2
      exit 1
      ;;
  esac
done

if [ -z "${TEMPO_ENDPOINT}" ]; then
  read -rsp 'Enter your Tempo endpoint (i.e., tempo-us-central1.grafana.net): ' TEMPO_ENDPOINT
  printf $'\n' >&2

  # We require a endpoint for the agent; we don't do this same check for
  # the username and password as the remote Tempo system may not have basic
  # auth enabled.
  if [ -z "${TEMPO_ENDPOINT}" ]; then
    echo "error: TEMPO_ENDPOINT must be provided by flag, env, or stdin" >&2
    exit 1
  fi
fi

if [ -z "${TEMPO_USERNAME}" ] && [ "${TEMPO_USERNAME_SET}" -eq 0 ]; then
  read -rsp 'Enter your Tempo username: ' TEMPO_USERNAME
  printf $'\n' >&2
fi

if [ -z "${TEMPO_PASSWORD}" ] && [ "${TEMPO_PASSWORD_SET}" -eq 0 ]; then
  read -rsp 'Enter your Tempo password: ' TEMPO_PASSWORD
  printf $'\n' >&2
fi

export NAMESPACE
export TEMPO_ENDPOINT
export TEMPO_USERNAME
export TEMPO_PASSWORD

curl -fsSL "$MANIFEST_URL" | envsubst
