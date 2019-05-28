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

# This script fetch list of current PRs, checks which of those are in a mergeable state and then it performs the merge
[[ ! -n ${GITHUB_TOKEN} ]] && echo "GITHUB_TOKEN env variable must be set" \
    && exit 1

api_url="https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/skuba"
pull_url="${api_url}/pulls"
issue_url="${api_url}/issues"
status_url="${api_url}/statuses"
commit_url="${api_url}/commits"
ci_integration_job="https://ci.suse.de/view/CaaSP/job/caasp-jobs/job/caasp-v4-integration/job"

# setup_credentials: Prepares the GIT_ASKPASS variable
setup_credentials() {
    local token_file="${WORKSPACE:-$(pwd)}/token.sh"

    # GIT_ASKPASS needs the name of the script which returns the credential we are going to use.
    # It's not pretty but it will get the job done
    cat > "${token_file}" <<EOF
#!/bin/bash
# If the GITHUB_TOKEN is a user:token pair, then we only want the token part of it
echo ${GITHUB_TOKEN} | cut -d ':' -f 2
EOF
    chmod a+x "${token_file}"
    export GIT_ASKPASS="${token_file}"
}

# cleanup_pr: Removes PR artifacts from GitHub
# $1: The PR number to remove
# $2: The testing CI branch on GitHub
# Returns: True if the artifacts were removed or false if they did not
cleanup_pr() {
    local pr=${1} pr_branch=${2}
    # Just remove the remote branch
    git push origin --delete ${pr_branch}
}

# cleanup_stale_prs: Removes CI branches for old PRs
# Returns: True if all branches has been removed or false if they did not
cleanup_stale_prs() {
    for pr_branch in $(git ls-remote origin | \
        grep -Eo "ci-test-pr-[[:digit:]]+"); do
        # Check that the PR was closed
        pr_number=$(echo $pr_branch | rev | cut -d "-" -f1 | rev)
        echo "Cleaning up $pr_branch branch for PR-${pr_number} if needed..."
        curl -s ${pull_url}/${pr_number} | jq -cr '. | .state' \
            | grep -q "closed" && git push origin --delete ${pr_branch}
    done
}

# merge_pr: Merges a GitHub PR
# $1: The PR number to merge
# Returns: True if the PR was merged, or False if it didn't
merge_pr() {
    local pr=${1} pr_ref

    echo "Merging PR-${pr}..."
    if curl -i -s -X PUT -H "Content-Type: application/json" \
        -d '{"merge_method":"merge"}' ${pull_url}/$pr/merge | grep Status \
        | grep -q 200; then
        echo "The PR-${pr} was merged!"
        return 0
    else
        echo "Failed to merge PR-${pr}"
        return 1
    fi
}

# rebased_pr_comment: Sends a GitHub comment to PR that rebase job failed
# $1: The PR number
# $2: The status to send
# Returns: True if the notification was sent or False if it didn't
rebase_pr_comment() {
    local pr=${1}

    if curl -s ${issue_url}/${pr}/comments -H "Content-Type: application/json"\
        -X PUT \
        -d "{\"body\":\"Rebased CI job failed: ${ci_integration_job}/ci-test-pr-${pr}/lastBuild\"}" | grep Status | grep -q 201; then
        echo "Status for rebased PR-${pr} was sent!"
        return 0
    else
        echo "Failed to send status for rebased PR-${pr}!"
        return 1
    fi
}

# check_pr_statuses: Obtains all the statuses for a PR.
# $1: The PR we are testing
# $2: The head ref for the given PR
# Returns: True if all the statuses are green or Failed if one is not
check_pr_statuses() {
    local pr=${1} pr_ref=${2}

    # Check if the PR has any failed check status
    if curl -s ${commit_url}/${pr_ref}/status | jq -rc '.state' | grep -vq success; then
        echo "PR-${pr} has some CI failures. Skipping..."
        return 1
    else
        return 0
    fi
}

# check_job_result: Checks the result of the last build for the integration job
# $1: The PR number for the job
# Returns: True if the job passed, or False if it failed.
check_job_result() {
    local pr=${1} status

    status=$(curl -s ${ci_integration_job}/ci-test-pr-${pr}/lastBuild/api/json | jq -cr '.result')
    [[ "${status}" == "SUCCESS" ]] && return 0
    return 1
}

# ci_update_pr: Create a CI test branch to test a GitHub PR
# $1: The PR number to test
# $2: Base branch to use for rebase
# Returns: True if the PR branch was succesfully created and rebased or if an existing one passed the CI
# and the PR was merged.
ci_update_pr() {
    local pr=${1} pr_branch="ci-test-pr-${1}" base_ref=${2}

    # Clean up some leftovers from before
    git branch -D ${pr_branch} || :
    git rebase --abort || :

    # Checkout the PR and create the branch
    git fetch origin pull/${pr}/head
    git checkout -f -b ${pr_branch} FETCH_HEAD

    # Fetch base branch. This populates FETCH_HEAD
    git fetch origin ${base_ref}

    # And now rebase it to bring it up to date
    git rebase FETCH_HEAD

    # If we have a branch upstream, we can compare it to see if the
    # developer pushed new changes to the PR which invalidated the testing branch.
    if git ls-remote -q origin | grep -q refs/heads/${pr_branch}; then
        if git diff --quiet ${pr_branch}..remotes/origin/${pr_branch}; then
            # If there is no diff, then we do not need to rebuild the branch
            if check_job_result ${pr}; then
                echo "${pr_branch} is updated and CI passed. Merging..."
                merge_pr ${pr}
                cleanup_pr ${pr} ${pr_branch}
            else
                echo "${pr_branch} is updated and CI failed. Aborting..."
                rebase_pr_comment ${pr}
            fi
            return 0
        fi
    fi
    # We are here because either this is a new branch or the PR was updated.
    echo "Force-pushing the updated ${pr_branch}"
    git push -f origin ${pr_branch}
}

# Wrapper around sleep call which is necessary whilst GH is re-calculating
# the PR status
need_sleep() {
    # FIXME: OK so this is a bit random but we need to give GH a few seconds
    # to update the status of pending PRs once one of them is merged. Maybe
    # we can improve that through API polling or something.
    sleep ${1}
}

# Prepare git credentials
setup_credentials

# Ensure git repo is configured correctly
git config user.email containers-bugowner@suse.de
git config user.name 'CaaSP CI'
git config user.user caaspjenkins

# Lets start with cleaning up closed PRs
cleanup_stale_prs

# Fetch a list of open PRs which are not labeled as 'wip' or 'do not merge'. The return format is
# $pr_number, $remote_repo, $remote_ref, $remote_sha, $base_repo, $base_ref
for pr_info in $(curl -s -X GET -H "Content-Type: application/json" -d '{"state":"open"}' ${pull_url} \
    | jq '[ .[] | select(.labels == [] or ((.labels | contains([{name: "wip"}]) | not) and (.labels | contains([{name: "do not merge"}]) | not))) ]' \
    | jq -rc '. | unique_by(.url)[] | [(.number | tostring), .head.repo.full_name, .head.ref, .head.sha, .base.repo.full_name, .base.ref] | join(",")'); do
    pr=$(echo $pr_info | awk -F, '{print $1}')
    head_repo=$(echo $pr_info | awk -F, '{print $2}')
    head_ref=$(echo $pr_info | awk -F, '{print $3}')
    head_sha=$(echo $pr_info | awk -F, '{print $4}')
    base_repo=$(echo $pr_info | awk -F, '{print $5}')
    base_ref=$(echo $pr_info | awk -F, '{print $6}')

    # Obtain the mergeable_state attribute so determine if the PR is good to get merged or not.
    merge_status=$(curl -s -X GET ${pull_url}/$pr | jq -rc '. | .mergeable_state')
    echo "Examining PR-${pr} (status: ${merge_status}, baseref=${base_ref})"

    # Skip PR if some of the CI tests have failed
    check_pr_statuses ${pr} ${head_sha} || continue

    # This is the meaning of the following GitHub statuses
    # clean: The PR passed CI and approved
    # behind: The PR passed the CI but it's not rebased
    # blocked: The PR either failed the CI tests or it does not have enough approvals
    # dirty: The PR has conflicts that need manual resolving
    case ${merge_status} in
        behind|clean)
            echo "PR-${pr} is now being rebased and merged..."
            ci_update_pr ${pr} ${base_ref}
            need_sleep 10
            ;;
        blocked)
            echo "PR-${pr} has not been approved. Skipping..."
            ;;
        dirty)
            echo "PR-${pr} has conflicts that need manual resolving. Skipping..."
            ;;
        *)
            echo "merge_status: '${merge_status}' is currently unhandled"
            continue
            ;;
    esac
done

exit 0
