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

add_package_to_need_reboot() {
    echo "$1" >> /etc/zypp/needreboot.d/"$1"
}

add_repository() {
    zypper ar -f --no-gpgcheck "/suse/artifacts/repos/$1" "$1"
}

refresh_repository() {
    zypper ref -f -r "$1"
}

install_package() {
    zypper -n in -r "${1:-base}" "${2:-caasp-test}"
}

zypper_show_patch() {
    zypper --no-refresh info -r "$1" -t patch "$2"
}

check_patch_type_interactivity() {
    [ "$(zypper_show_patch "$1" "$2" | grep -Poh '(?<=Interactive : )([^\s]+)')" == "$3" ]
}

check_list_patches_interactivity() {
    [ "$(zypper --non-interactive --xmlout list-patches | grep -Poh '(?<=interactive\=\")([a-zA-Z0-9]+)')" == "$1" ]
}

zypper_patch() {
    if [ "$SKUBA" = "1" ]; then
        if [ -n "$SKUBA_UPDATE_OPTIONS" ]; then
            skuba-update "$SKUBA_UPDATE_OPTIONS"
        else
            skuba-update
        fi
    elif [ "$SKUBA_UPDATE_OPTIONS" != "--annotate-only" ]; then
        zypper --no-refresh --non-interactive-include-reboot-patches patch -r "$1" -y
    fi
}

check_return_code() {
    if [ "$SKUBA" = "1" ]; then
        if [[ $1 -ne 0 ]]; then
            echo "unexpected return value ($1) from skuba-update: 0 expected"
            exit 1
        fi
    else
        if [[ $1 -ne $2 ]]; then
            echo "unexpected return value ($1) from zypper patch (expected $3: $2)"
            exit 1
        fi
    fi
}

check_test_package_version() {
    caasp-test | grep "CaaSP test version $1"
}

check_reboot_needed_present() {
    set +e
    zypper needs-rebooting
    ret=$?
    set -e
    [ "$ret" = "102" ] && [ -f /var/run/reboot-needed ]
}

check_reboot_required_present() {
    if [ "$SKUBA" = "1" ]; then
        [ -f /var/run/reboot-required ]
    fi
}

check_reboot_needed_absent() {
    set +e
    zypper needs-rebooting
    ret=$?
    set -e
    [ "$ret" != "102" ] && [ ! -f /var/run/reboot-needed ]
}

check_reboot_required_absent() {
    if [ "$SKUBA" = "1" ]; then
        [ ! -f /var/run/reboot-required ]
    fi
}

force_machine_id_file() {
    [ -w "/etc/machine-id" ] || echo "49f8e2911a1449b7b5ef2bf92282909a" >> /etc/machine-id
}

install_fixtures() {
    cp /suse/fixtures/fake_kubectl.sh /usr/bin/kubectl
    chmod +x /usr/bin/kubectl
}

check_kubectl_calls() {
    if [ "$SKUBA" = "1" ]; then
        for i in "$@"; do
            echo "$i" >> /commands.txt
        done
        type -p diff || zypper -n in diffutils
        diff -w /commands.txt /tmp/kubectl-commands &> /dev/null
    fi
}

check_no_kubectl_calls() {
    if [ "$SKUBA" = "1" ]; then
        [ ! -f /tmp/kubectl-commands ]
    fi
}
