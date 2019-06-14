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

FROM registry.opensuse.org/opensuse/leap:latest
MAINTAINER SUSE Containers Team <containers@suse.com>

COPY *requirements.txt /
COPY test/unit-tests-init.sh /init

RUN zypper -n ref && \
    zypper -n in python3-setuptools python3-pip ShellCheck lsof && \
    pip install -r /dev-requirements.txt && \
    chmod +x /init && \
    useradd -m -d /home/suse -s /bin/bash suse && \
    zypper clean -a

USER suse
ENTRYPOINT ["/init"]
