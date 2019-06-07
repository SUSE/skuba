#!/bin/bash

# Copyright 2019 SUSE LINUX GmbH, Nuernberg, Germany.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

[[ ! -n ${GITHUB_TOKEN} ]] && echo "GITHUB_TOKEN env variable must be set" && exit 1
[[ ! -n ${CHANGE_ID} ]] && echo "CHANGE_ID env variable must be set" && exit 1
[[ ! -n ${SUBDIRECTORY} ]] && echo "SUBDIRECTORY env variable must be set" && exit 1

# We only expect PRs to come from forked repositories instead of branches from the main repo
# so we need to check that before moving forward to examine the individual commits
if $(curl -sL https://patch-diff.githubusercontent.com/raw/SUSE/skuba/pull/${CHANGE_ID}.patch?token=${GITHUB_TOKEN} | \
    grep -q "diff --git a/${SUBDIRECTORY}/"); then
    echo "PR-${CHANGE_ID} contains changes in the ${SUBDIRECTORY} subfolder."
else
    echo "PR-${CHANGE_ID} does not contain changes in the ${SUBDIRECTORY} subfolder."
fi