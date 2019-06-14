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

RPMBUILD=/root/rpmbuild
PACKAGEDIR=/suse/specs
OUTPUTDIR=/suse/artifacts

build_init() {
    rpm_setuptree
}

rpm_setuptree() {
    rpmdev-setuptree
}

build_package() {
    rpmbuild -ba $PACKAGEDIR/"${1:-caasp-test-1-1.noarch}".spec
}

copy_packages() {
    find $RPMBUILD/RPMS -type f -name "*.rpm" -exec cp {} $OUTPUTDIR \;
}
