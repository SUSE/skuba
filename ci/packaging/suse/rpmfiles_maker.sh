#!/bin/bash

set -e

susepkg_dir=$(cd "$( dirname "$0" )" && pwd)
tmp_dir=$(mktemp -d -t skuba_XXXX)
rpm_files="${susepkg_dir}/obs_files"
version="$1"
tag="$2"
closest_tag="$3"

log()   { (>&2 echo ">>> $*") ; }
clean() { log "Cleaning temporary directory ${tmp_dir}"; rm -rf "${tmp_dir}"; }

if [ $# -ne 3 ]; then
    log "usage: <version> <tag> <closest-tag>"
    exit 1
fi

trap clean ERR

rm -rf "${rpm_files}"
mkdir -p "${rpm_files}"
git archive --prefix=skuba/ -o "${tmp_dir}/skuba.tar.gz" HEAD
sed -e "s|%%VERSION|${version}|;s|%%TAG|${tag}|;s|%%CLOSEST_TAG|${closest_tag}|" "${susepkg_dir}/skuba_spec_template" \
    > "${tmp_dir}/skuba.spec"
make CHANGES="${tmp_dir}/skuba.changes.append" suse-changelog
cp "${tmp_dir}"/* "${rpm_files}"
log "Find files for RPM package in ${rpm_files}"

ibs_user=$(osc config https://api.suse.de user | awk '$0=$NF' || echo -n '')
if [[ -n "$ibs_user" ]]
then
    log "Found IBS config; updating IBS"
    ibs_user=${ibs_user//\'}
    branch_project="home:${ibs_user}:caasp_auto_release"
    branch_name="skuba_$tag"
    work_dir="$tmp_dir/ibs_skuba"
    log "Creating IBS branch"
    osc -A 'https://api.suse.de' branch Devel:CaaSP:4.0 skuba \
      "$branch_project" "$branch_name"
    osc -A 'https://api.suse.de' co -o "$work_dir" \
      "$branch_project/$branch_name"
    log "Updating IBS branch"
    cp -v "$rpm_files/skuba.spec" "$rpm_files/skuba.tar.gz" "$work_dir/"
    cat "$rpm_files/skuba.changes.append" \
      "$work_dir/skuba.changes" \
      > "$tmp_dir/merged.changes" \
      && \
      cp -v "$tmp_dir/merged.changes" "$work_dir/skuba.changes"
    osc -A 'https://api.suse.de' ci "$work_dir" \
      -m "$(<"$rpm_files/skuba.changes.append")"
    log "Creating self-cleaning SR"
    osc -A 'https://api.suse.de' sr \
      -m "Update for release '$tag'" \
      --cleanup --yes \
      "$branch_project" "$branch_name" \
      Devel:CaaSP:4.0 skuba
fi

clean
