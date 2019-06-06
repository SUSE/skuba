#!/usr/bin/env bash

# Copyright (c) 2019 SUSE LLC. All rights reserved.
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

UPDATE_REPO="update-with-message"

add_repository "base"
install_package "base" "caasp-test"

check_test_package_version "1"

add_repository "$UPDATE_REPO"
zypper_show_patch "$UPDATE_REPO" "SUSE-2019-0"
check_patch_type_interactivity "$UPDATE_REPO" "SUSE-2019-0" "message"
zypper_patch "$UPDATE_REPO"

check_test_package_version "1"
check_reboot_needed_absent
check_reboot_required_absent
