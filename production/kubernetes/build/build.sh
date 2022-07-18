#!/usr/bin/env bash
# shellcheck shell=bash

set +e

DIRNAME=$(dirname "$0")

pushd "${DIRNAME}" || exit 1
# Make sure dependencies are up to date
jb install
tk show --dangerous-allow-redirect ./templates/bare > "${PWD}/../agent-bare.yaml"
tk show --dangerous-allow-redirect ./templates/loki > "${PWD}/../agent-loki.yaml"
tk show --dangerous-allow-redirect ./templates/traces > "${PWD}/../agent-traces.yaml"
tk show --dangerous-allow-redirect ./templates/operator > "${PWD}/../../operator/templates/agent-operator.yaml"
popd || exit 1
