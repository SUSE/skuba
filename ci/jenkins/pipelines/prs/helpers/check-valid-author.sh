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

# Ensure that all SUSE employees are using the correct email address.

[[ ! -n ${GITHUB_TOKEN} ]] && echo "GITHUB_TOKEN env variable must be set" && exit 1
GIT_BRANCH=${GIT_BRANCH:-origin/master}

# Check all commits in PR for SUSE email address
if ! git log --format=%ae --no-merges origin/master..HEAD | sort -n | uniq | grep -v -q '@suse\.\(com\|cz\|de\)'; then
	echo "All commits are from SUSE employees. Skipping further author checks..."
	exit 0
fi

for commit in $(git log --format=%H --no-merges ${GIT_BRANCH}..HEAD); do
	author_email=$(curl -s https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/caaspctl/commits/$commit | jq -cr '. | .author.login, .commit.author.email' | tr -d '"')
	login=$(echo $author_email | awk '{print $1}')
	author=$(echo $author_email | awk '{print $2}')
	echo "Checking if $login($author) is part of the SUSE organization"
	if curl -i -s https://${GITHUB_TOKEN}@api.github.com/orgs/SUSE/members/$login | grep Status | grep -q 204; then
		echo "$login($author) is part of SUSE organization but a SUSE e-mail address was not used in commit: $commit"
		exit 1
	fi
done


