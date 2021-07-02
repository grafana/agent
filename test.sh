#!/usr/bin/env bash

if [ ${BUILD_IN_CONTAINER} = "false" ]
then
    echo true
else
    echo false
fi