#!/bin/bash 

set -e

pkg_dir="ci/packaging/suse"
sdist_dir="${pkg_dir}/obs_files"
changes="${sdist_dir}/skuba-update.changes.append"

log()   { (>&2 echo ">>> $*") ; }
clean() { log "Cleaning ${sdist_dir}"; rm -rf "${sdist_dir}"; }

trap clean ERR

ver_str=$(grep "version =" setup.py)
ver_str=${ver_str#*\'}
ver=${ver_str%%\'*}
sum_str=$(grep "description =" setup.py)
sum_str=${sum_str#*\"}
sum=${sum_str%%\"*}

desc_str=$(grep "long_description =" setup.py)
desc_str=${desc_str#*\"}
desc=${desc_str%%\"*}


[ -d "${sdist_dir}" ] && clean

./setup.py sdist --dist-dir "${sdist_dir}"
mv "${sdist_dir}/skuba-update-${ver}.tar.gz" "${sdist_dir}/skuba-update.tar.gz"
sed -e "s|^Version:\(.*\)__VERSION__.*$|Version:\1${ver}|g" \
    -e "s|^Summary:\(.*\)__SUMMARY__.*$|Summary:\1${sum}|g" \
    -e "s|__DESCRIPTION__|${desc}|g" \
    "${pkg_dir}/skuba-update_spec.tmpl" > "${sdist_dir}/skuba-update.spec"
make CHANGES="${changes}" suse-changelog
log "Find files for RPM package in ${sdist_dir}"
