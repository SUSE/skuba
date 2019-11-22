#!/bin/bash

dashes="-------------------------------------------------------------------"
header="${dashes}%n%cd - %an <%ae>"
datef="%a %b %e %H:%M:%S %Z %Y"
changes=${1:-skuba.changes.append}

[ -f "${changes}" ] && rm "${changes}"

mapfile -t tags < <( git tag | sort -rV | head -n2 )
scope="${tags[1]}...${tags[0]}"

{
    git show --format="${header}%n%n- Update to ${tags[0]}:" \
        --date="format-local:${datef}" -s "${tags[0]}^{commit}"
    git log -s --cherry-pick --format="%w(77,2,12)* %h %s" --no-merges "${scope}"
    # Add empty line
    echo
} >> "${changes}"
