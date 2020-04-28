#!/bin/bash

set -e
fileversion=$(cat skuba-update/VERSION)
gitversion=$(git describe --tags | sed -E 's/v(([0-9]\.?)+).*/\1/')
if [ "$fileversion" != "$gitversion" ]; then
    echo "incorrect python version file, please run make"
    exit 1
fi
