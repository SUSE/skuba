#!/usr/bin/env bash

# Copyright (c) 2019 SUSE LLC.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -xe

WORKDIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
PROJECT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." &>/dev/null && pwd)

if [ -z "$TESTNAME" ]; then
    tests="$WORKDIR/tests/*.sh"
else
    tests="$WORKDIR"/tests/"$TESTNAME".sh
fi

if [ "$SKUBA" != "1" ]; then
    for name in $tests
    do
        docker run --rm -v "$WORKDIR":/suse -v "$PROJECT":/usr/src "$1" /suse/tests/"$(basename "$name")"
    done
fi

if [ "$SKUBA" = "1" ] || [ -z "$SKUBA" ]; then
    # Add python into the original image and use the resulting one from now on.
    dst="$1-python"
    random=$(head -c 500 /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 10 | head -n 1)
    tmpname="suse-python-builder-$random"
    docker run -v "$WORKDIR":/suse -v "$PROJECT":/usr/src --name "$tmpname" "$1" /bin/bash /suse/install_skuba_update.sh
    docker commit "$tmpname" "$dst"
    docker rm -f "$tmpname"

    # Execute skuba-update in the integration tests instead of zypper patch.
    for name in $tests
    do
        docker run --rm -v "$WORKDIR":/suse -v "$PROJECT":/usr/src -e SKUBA=1 "$dst" /suse/tests/"$(basename "$name")"
    done
fi
