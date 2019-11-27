#!/bin/bash

set -e
for file in "$@"; do
    go fmt "$file"
done
