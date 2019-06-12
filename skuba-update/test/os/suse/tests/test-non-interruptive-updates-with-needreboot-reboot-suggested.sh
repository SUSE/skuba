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

source "$(dirname "$0")/../suse.sh"

UPDATE_REPO="update-with-reboot-suggested"

force_machine_id_file
install_fixtures

add_repository "base"
install_package "base" "caasp-test"

check_test_package_version "1"

add_package_to_need_reboot "caasp-test"

add_repository "$UPDATE_REPO"
set +e
zypper_show_patch "$UPDATE_REPO" "SUSE-2019-0"
check_patch_type_interactivity "$UPDATE_REPO" "SUSE-2019-0" "reboot"
zypper_patch "$UPDATE_REPO"
zypper_retval=$?
set -e

check_return_code $zypper_retval 102 "ZYPPER_EXIT_INF_REBOOT_NEEDED"

check_test_package_version "2"
check_reboot_needed_present
check_reboot_required_present

check_kubectl_calls "kubectl get nodes -o json" \
                    "kubectl annotate --overwrite node my-node-1 caasp.suse.com/has-updates=yes" \
                    "kubectl annotate --overwrite node my-node-1 caasp.suse.com/has-security-updates=no" \
                    "kubectl annotate --overwrite node my-node-1 caasp.suse.com/has-disruptive-updates=no"
