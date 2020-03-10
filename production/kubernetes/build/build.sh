#!/usr/bin/env bash
set +e

# Make sure dependencies are up to date
jb install
tk show --dangerous-allow-redirect ./template > ../agent.yaml

