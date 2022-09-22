#!/bin/bash

# This script generates runt.sh script that pulls down the needed files and runs them.

echo "#!/bin/bash" > runt.sh

# Instead of `for find .` doing it this way due to https://www.shellcheck.net/wiki/SC2044.
while IFS= read -r -d '' i
do
    # Ignore current directory, png and ds_store files. 
    if [[ $i == "." || $i == "./.DS_Store" || $i == *.png || $i == *.sh ]];
    then
        continue
    fi
    # If this is a directory create the directory ignoring if it already exists (-p).
    if [ -d "$i" ];
    then
        echo "mkdir -p $i" >> runt.sh
    else
        # Trim the '.' off the beginning, the file is './assets/file.flow' and need to remove '.'.
        trimName="${i:1}"
        # TODO at some point change this to release.
        echo "curl https://raw.githubusercontent.com/grafana/agent/main/docs/sources/flow/tutorials/assets$trimName -o $i" >> runt.sh
    fi
done <   <(find . -print0)

# Always pull the newest.
# TODO at some point change this from main.
echo "docker pull grafana/agent:main " >> runt.sh
echo "CONFIG_FILE=\$1 docker-compose -f ./docker-compose.yaml up" >> runt.sh
