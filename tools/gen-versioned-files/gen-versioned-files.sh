#!/bin/sh
AGENT_VERSION=$(cat ./tools/gen-versioned-files/agent-version.txt | tr -d '\n')

if [ -z "$AGENT_VERSION" ]; then
    echo "AGENT_VERSION can't be found. Are you running this from the repo root?"
    exit 1
fi

versionMatcher='^v[0-9]+\.[0-9]+\.[0-9]+(-rc[0-9]+)?$'

if ! echo "$AGENT_VERSION" | grep -Eq "$versionMatcher"; then
    echo "AGENT_VERSION env var is not in the correct format. It should be in the format of vX.Y.Z or vX.Y.Z-rcN"
    exit 1
fi

templates=$(find . -type f -name "*.t.*" -not -path "./.git/*")
for template in $templates; do
    # Extract the original file extension
    file_extension="${template##*.}"

    # Extract the file name without the extension
    file_name_without_ext="${template%.*}"
    file_name_without_t="${file_name_without_ext%.*}"

    # Construct the new file path by the extension to the stripped file name
    new_file="${file_name_without_t}.${file_extension}"
    echo "Generating $new_file"
    sed -e "s/\$AGENT_VERSION/$AGENT_VERSION/g" < "$template" > "$new_file"
done
