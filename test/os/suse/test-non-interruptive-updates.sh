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

# Non-interruptive updates are those that are not reported as a reboot suggested in the
# update metadata. Zypper will flag a 0 return code in this case, and will not
# write the /var/run/reboot-needed sentinel file.

source "$(dirname "$0")/suse.sh"

add_repository "base"
install_package "base" "caasp-test"

check_test_package_version "1"

add_repository "update-without-reboot-suggested"
zypper_patch "update-without-reboot-suggested"

check_test_package_version "2"
check_reboot_needed_absent
