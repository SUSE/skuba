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

# Interruptive updates are those that are reported as a reboot suggested in the
# update metadata. Zypper will flag a 102 return code in this case, and will
# write the /var/run/reboot-needed sentinel file.

source "$(dirname "$0")/suse.sh"

add_repository "base"
install_package "base" "caasp-test"

check_test_package_version "1"

add_package_to_need_reboot "caasp-test"

add_repository "update-with-reboot-suggested"
set +e
zypper_patch "update-with-reboot-suggested"
zypper_retval=$?
set -e

if [[ $zypper_retval -ne 102 ]]; then
    echo "unexpected return value ($zypper_retval) from zypper patch (expected ZYPPER_EXIT_INF_REBOOT_NEEDED: 102)"
    exit 1
fi

check_test_package_version "2"
check_reboot_needed_present
