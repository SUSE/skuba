#!/bin/bash
##########################################
# generate skuba tar.gz and specfile
##########################################

##########################################
# Housekeeping
set -e
tmp_dir=$(mktemp -d -t skuba_XXXX)
log()   { (>&2 echo ">>> $*") ; }
clean() { log "Cleaning temporary directory ${tmp_dir}"; rm -rf "${tmp_dir}"; }
trap clean ERR
susepkg_dir=$(cd "$( dirname "$0" )" && pwd)
rpm_files="${susepkg_dir}/obs_files"
devel_prj="Devel:CaaSP:5"

##########################################
# CLI args
if (( $# != 3 )); then
    log "usage: <version> <tag> <closest-tag>"
    exit 1
fi
version="$1"     # pretty version string for display 
tag="$2"         # can be ugly refs/tags/xxx or specific commit hash
closest_tag="$3" # usually $tag, but might be commit

##########################################
# env vars which come from GitHub Actions
TARBALL_PATH=${TARBALL_PATH:-${tmp_dir}/skuba.tar.gz}
TEMPLATE_PATH=${TEMPLATE_PATH:-${susepkg_dir}/skuba_spec_template}
SPECFILE_PATH=${SPECFILE_PATH:-${tmp_dir}/skuba.spec}

##########################################
# other global stuff

function main {
    gen_tar
    gen_spec
    if [[ "${GITHUB_ACTIONS:-}" != "true" ]]
    then
        # GitHub Actions calls this script directly
        make CHANGES="${tmp_dir}/skuba.changes.append" suse-changelog

        # Assemble for OBS / manual use later 
        rm -rf "${rpm_files}"
        mkdir -p "${rpm_files}"
        cp "${tmp_dir}"/* "${rpm_files}"
        log "Find files for RPM package in ${rpm_files}"

        ibs_user=$(osc config https://api.suse.de user | awk '$0=$NF' \
                   || echo -n '')
        if [[ -n "$ibs_user" ]]
        then
            do_ibs_update "$ibs_user"
        fi
    fi
}

# generate tarball
function gen_tar {
    log "Generating tarball at '$TARBALL_PATH'"
    git archive --prefix=skuba/ -o "$TARBALL_PATH" "$tag"
}

# generate specfile
function gen_spec {
    log "Generating specfile at '$SPECFILE_PATH'"
    sed -e "
      s|%%VERSION|${version}|;
      s|%%TAG|${tag}|;
      s|%%CLOSEST_TAG|${closest_tag}|
      " "$TEMPLATE_PATH" > "$SPECFILE_PATH"
}

function do_ibs_update {
    log "Found IBS config; updating IBS"
    user=${1//\'}
    branch_project="home:${user}:caasp_auto_release"
    branch_name="skuba_$tag"
    work_dir="$tmp_dir/ibs_skuba"
    log "Creating IBS branch"
    osc -A 'https://api.suse.de' branch --force "${devel_prj}" skuba \
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
    log "Running OBS services"
    (
        cd "$work_dir"
        osc -A 'https://api.suse.de' service disabledrun
    )
    osc -A 'https://api.suse.de' ci "$work_dir" \
      -m "$(<"$rpm_files/skuba.changes.append")"
}

main

clean
