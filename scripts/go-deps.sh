#!/bin/sh
################################################################################
# Script to set replace() lines up for using kubernetes as a library, despite
#  guidance to the contrary from upstream. :) Mostly stolen from:
# https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-521493597
#
# Usage: $0 1.16.2
# then go through and clean up the replace stanza to be one consistent block
################################################################################
set -euo pipefail

cd "$( dirname "$0")"/..

VERSION=${1#"v"}
if [ -z "$VERSION" ]; then
    echo "Must specify version!"
    exit 1
fi
MODS=($(
    curl -sS https://raw.githubusercontent.com/kubernetes/kubernetes/v${VERSION}/go.mod |
    sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
))
for MOD in "${MODS[@]}"; do
    V=$(
        go mod download -json "${MOD}@kubernetes-${VERSION}" |
        sed -n 's|.*"Version": "\(.*\)".*|\1|p'
    )
    go mod edit "-replace=${MOD}=${MOD}@${V}"
done
