#!/bin/bash

set -e

mkfile_dir=$(pwd)
susepkg_dir=$(cd "$( dirname "$0" )" && pwd)
tmp_dir=$(mktemp -d -t caaspctl_XXXX)
rpm_files="${susepkg_dir}/obs_files"
version=$1

log()   { (>&2 echo ">>> $*") ; }
clean() { log "Cleaning temporary directory ${tmp_dir}"; rm -rf "${tmp_dir}"; }

if [ $# -ne 1 ]; then
    log "missing caaspctl package version parameter"
    exit 1
fi

trap clean ERR

rm -rf "${rpm_files}"
mkdir -p "${rpm_files}"
tar --exclude=".*" -caf "${tmp_dir}/caaspctl.tar.gz" \
    -C "$(dirname "${mkfile_dir}")" caaspctl
sed -e s"|%%VERSION|${version}|" "${susepkg_dir}/caaspctl_spec_template" \
    > "${tmp_dir}/caaspctl.spec"
make CHANGES="${tmp_dir}/caaspctl.changes.append" suse-changelog
cp "${tmp_dir}"/* "${rpm_files}"
log "Find files for RPM package in ${rpm_files}"
clean
