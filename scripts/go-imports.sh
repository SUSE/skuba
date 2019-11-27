#!/bin/bash

set -e
for file in "$@"; do
    goimports -l -w --local=github.com/SUSE "$file"
done
