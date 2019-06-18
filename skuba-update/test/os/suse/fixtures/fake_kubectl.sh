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

if [ "$1" = "get" ]; then
    cat <<EOF
{
    "items": [
        {
            "metadata": {
                "name": "my-node-1"
            },
            "status": {
                "nodeInfo": {
                    "machineID": "49f8e2911a1449b7b5ef2bf92282909a"
                }
            }
        },
        {
            "metadata": {
                "name": "my-node-2"
            },
            "status": {
                "nodeInfo": {
                    "machineID": "9ea12911449eb7b5f8f228294bf9209a"
                }
            }
        }
    ]
}
EOF
fi

echo "kubectl" "$@" >> /tmp/kubectl-commands
