#!/usr/bin/env bash
set +e

DIRNAME=$(dirname $0)

pushd ${DIRNAME}
# Make sure dependencies are up to date
jb install
tk show --dangerous-allow-redirect ./template > ${PWD}/../agent.yaml
popd
