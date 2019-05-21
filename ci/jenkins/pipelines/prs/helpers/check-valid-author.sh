#!/bin/bash

# Copyright 2019 SUSE LINUX GmbH, Nuernberg, Germany.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

[[ ! -n ${GITHUB_TOKEN} ]] && echo "GITHUB_TOKEN env variable must be set" && exit 1
[[ ! -n ${CHANGE_ID} ]] && echo "CHANGE_ID env variable must be set" && exit 0

# We only expect PRs to come from forked repositories instead of branches from the main repo
# so we need to check that before moving forward to examine the individual commits
if [[ $(curl -s https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/caaspctl/pulls/${CHANGE_ID} | \
    jq -rc '.| if (.head.repo.full_name == .base.repo.full_name) then true else false end') == true ]]; then
    echo "PR-${CHANGE_ID} is coming from a branch in the target repository. This is not allowed!"
    echo "Please send your PR from a forked repository instead."
    exit 1
fi

# This for loop uses the GitHub API to fetch all commits in a PR and outputs the information in the following format
# $sha,$github_username,$author_email_address. If the author is using a SUSE address, then no further checks are necessary and we
# check the next commit. If the author is not using a SUSE email address, then we check if the user belongs to the SUSE organization.
# If he/she does, then we exit with non-zero exit code to denote that the user must be using a SUSE email address if he/she is a
# SUSE employee.
for commit_author in $(curl -s https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/caaspctl/pulls/${CHANGE_ID}/commits | jq -cr '.[] | [.sha, .author.login, .commit.author.email] | join(",")'); do
    commit=$(echo $commit_author | awk -F, '{print $1}')
    login=$(echo $commit_author | awk -F, '{print $2}')
    author=$(echo $commit_author | awk -F, '{print $3}')
    echo $author | grep -q '@suse\.\(com\|cz\|de\)' && echo "commit $commit is from SUSE employee $login($author). Moving on..." && continue
    echo "Checking if $login($author) is part of the SUSE organization"
    if curl -i -s https://${GITHUB_TOKEN}@api.github.com/orgs/SUSE/members/$login | grep Status | grep -q 204; then
        echo "$login($author) is part of SUSE organization but a SUSE e-mail address was not used in commit: $commit"
        exit 1
    fi
done
