#!/bin/bash

# if we're in a Github Action ($GITHUB_ACTIONS=true)
# and GIT_WORK_TREE isn't set, ensure git commands work...
if [[ -n "${GITHUB_ACTIONS:-}" && -z "${GIT_WORK_TREE=:-}" ]]
then
  export GIT_WORK_TREE=$GITHUB_WORKSPACE
fi

dashes=$( printf -- "-%.0s" {1..67} ) # why 67 hyphens? Who knows.
header="${dashes}%n%cd - %an <%ae>"
datef="%a %b %e %H:%M:%S %Z %Y"

changes=${1:-skuba.changes.append}
# default to putting release body in same location as $changes file
touch "$changes"
release_body=${RELEASE_BODY:-$(find "$changes" -printf '%h/release_body.txt\n')}

mapfile -t tags < <( git tag | sort -rV | head -n2 )
cur_tag=${CURRENT_TAG:-${tags[0]}}
prv_tag=${PREV_TAG:-${tags[1]}}
pretty_tag=${PRETTY_TAG:-${cur_tag##*/}} # make sure no "refs/tags/v..." format

function bulleted_changelog {
    scope=$1
    git log \
      --no-patch \
      --no-merges \
      --cherry-pick \
      --format="%w(77,2,12)* %h %s" \
      "$scope"
}

##########################################
# Generate GitHub release body
{
    echo "Update to $pretty_tag"
    bulleted_changelog "${prv_tag}...${cur_tag}"
} > "${release_body}"

##########################################
# Generate OBS changelog file
{
    git show --format="${header}%n%n- Update to ${cur_tag}:" \
        --date="format-local:${datef}" -s "${cur_tag}^{commit}"
    bulleted_changelog "${prv_tag}...${cur_tag}"
    # Add empty line
    echo ""
} > "${changes}"
