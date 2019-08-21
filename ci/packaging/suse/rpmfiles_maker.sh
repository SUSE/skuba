#!/bin/bash

set -e

susepkg_dir=$(cd "$( dirname "$0" )" && pwd)
tmp_dir=$(mktemp -d -t skuba_XXXX)
rpm_files="${susepkg_dir}/obs_files"
version="$1"
tag="$2"
annotated_tag="$3"

log()   { (>&2 echo ">>> $*") ; }
clean() { log "Cleaning temporary directory ${tmp_dir}"; rm -rf "${tmp_dir}"; }

if [ $# -ne 3 ]; then
    log "usage: <version> <tag> <annotated-tag>"
    exit 1
fi

trap clean ERR

rm -rf "${rpm_files}"
mkdir -p "${rpm_files}"
git archive --prefix=skuba/ -o "${tmp_dir}/skuba.tar.gz" HEAD
sed -e "s|%%VERSION|${version}|;s|%%TAG|${tag}|;s|%%ANNOTATED_TAG|${annotated_tag}|" "${susepkg_dir}/skuba_spec_template" \
    > "${tmp_dir}/skuba.spec"
make CHANGES="${tmp_dir}/skuba.changes.append" suse-changelog
cp "${tmp_dir}"/* "${rpm_files}"
log "Find files for RPM package in ${rpm_files}"
clean
