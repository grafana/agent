#!/bin/bash

if [ -z "$AGENT_VERSION" ]; then
    echo "AGENT_VERSION env var is not set"
    exit 1
fi

templates=$(find . -type f -name "*.templ" -not -path "./.git/*")
for template in $templates; do
    echo "Generating ${template%.templ}"
    envsubst < $template > ${template%.templ}
done
