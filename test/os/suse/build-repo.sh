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

source "$(dirname "$0")/suse_build.sh"

REPONAME=$1
REPODIR=/suse/artifacts/repos
REPOPATH=$REPODIR/$1

mkdir -p "$REPOPATH"

populate_repo() {
    case "$REPONAME" in
        base)
            cp /suse/artifacts/caasp-test-1-1.noarch.rpm "$REPOPATH"
            ;;
        update-*)
            cp /suse/artifacts/caasp-test-2-1.noarch.rpm "$REPOPATH"
            ;;
        *)
            echo "unknown repo ($REPONAME) requested"
            exit 1
    esac
}

initialize_repo() {
    createrepo --update "$REPOPATH"
}

add_erratum_to_repository() {
    cp /suse/repos/"$REPONAME".xml "$REPOPATH"/repodata/updateinfo.xml
    updateinfosha256=$(sha256sum "$REPOPATH"/repodata/updateinfo.xml | awk '{print $1}')
    updateinfosize=$(du -b "$REPOPATH"/repodata/updateinfo.xml | awk '{print $1}')
    gzip "$REPOPATH"/repodata/updateinfo.xml
    updateinfosha256gz=$(sha256sum "$REPOPATH"/repodata/updateinfo.xml.gz | awk '{print $1}')
    updateinfosizegz=$(du -b "$REPOPATH"/repodata/updateinfo.xml.gz | awk '{print $1}')
    mv "$REPOPATH"/repodata/updateinfo.xml.gz "$REPOPATH"/repodata/"$updateinfosha256gz"-updateinfo.xml.gz
    repodata=$(head -n-1 "$REPOPATH"/repodata/repomd.xml)
    echo "$repodata" > "$REPOPATH"/repodata/repomd.xml
    cat <<EOF >> "$REPOPATH"/repodata/repomd.xml
<data type="updateinfo">
  <checksum type="sha256">$updateinfosha256gz</checksum>
  <open-checksum type="sha256">$updateinfosha256</open-checksum>
  <location href="repodata/$updateinfosha256gz-updateinfo.xml.gz"/>
  <timestamp>$(date +%s)</timestamp>
  <size>$updateinfosizegz</size>
  <open-size>$updateinfosize</open-size>
</data>
</repomd>
EOF
}

populate_repo
initialize_repo

case "$REPONAME" in
    update-*)
        add_erratum_to_repository
    ;;
esac
