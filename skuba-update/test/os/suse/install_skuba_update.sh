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

release=$(head -n 1 /etc/os-release)
if [ "$release" = "NAME=\"SLES\"" ]; then
    rpm -e container-suseconnect
    zypper ar --no-gpgcheck http://download.opensuse.org/distribution/leap/15.1/repo/oss/ leap15.1
    zypper ar --no-gpgcheck http://download.opensuse.org/update/leap/15.1/oss updates
    zypper ref
fi

zypper in -y python3-setuptools

if [ "$release" = "NAME=\"SLES\"" ]; then
    zypper rr leap15.1
    zypper rr updates
fi

cd /usr/src
python3 setup.py install --root / --install-script /usr/sbin

exit 0
