#!/bin/sh

if [ -z "$AGENT_VERSION" ]; then
    echo "AGENT_VERSION env var is not set"
    exit 1
fi

versionMatcher='^v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?$'

if ! echo "$AGENT_VERSION" | grep -Eq "$versionMatcher"; then
    echo "AGENT_VERSION env var is not in the correct format. It should be in the format of vX.Y.Z or vX.Y.Z-rcN"
    exit 1
fi

templates=$(find . -type f -name "*.t" -not -path "./.git/*")
for template in $templates; do
    echo "Generating ${template%.t}"
    sed -e "s/\$AGENT_VERSION/$AGENT_VERSION/g" < "$template" > "${template%.t}"
done