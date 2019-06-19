import time

import jenkins
from github import GithubException


class PrMerge:
    def __init__(self, jenkins_config, repo):
        self.jenkins_config = jenkins_config
        self.repo = repo

    def merge_prs(self):
        mergeable_prs = PrMerge._get_mergeable_prs(self.repo)

        for mergeable_pr in mergeable_prs:
            if self._passed_integration_tests(mergeable_pr):
                try:
                    PrMerge._merge_pr(mergeable_pr)
                except GithubException as ex:
                    print(f'Merging PR-{mergeable_pr.number} failed due to:\n{ex}')

            # Give GitHub a break
            time.sleep(5)

    def _passed_integration_tests(self, pull_request):
        job_name = f'caasp-jobs/caasp-v4-test/PR-{pull_request.number}'
        wait_count = 15
        server = jenkins.Jenkins(self.jenkins_config['jenkins']['url'],
                                 username=self.jenkins_config['jenkins']['user'],
                                 password=self.jenkins_config['jenkins']['password'])
        job = server.get_job_info(job_name)
        next_build_number = job['nextBuildNumber']
        next_build_info = None

        print(f'Starting build {next_build_number} for {job_name}')

        server.build_job(job_name)

        # Wait for job to start
        while wait_count > 0:
            time.sleep(60)
            try:
                next_build_info = server.get_build_info(job_name, next_build_number)
            except jenkins.JenkinsException as ex:
                print(ex)
                print('Job waiting to start...')
            else:
                break

            wait_count -= 1

        if next_build_info is None:
            raise Exception("Job still hasn't started exiting!")

        print('Job started waiting for it to finish...')

        # Wait for job to finish
        while next_build_info['building']:
            time.sleep(10)
            next_build_info = server.get_build_info(job_name, next_build_number)

        print(f'Job finished with result {next_build_info["result"]}')

        return next_build_info['result'] == 'SUCCESS'

    @staticmethod
    def _not_wip_and_merge_allowed(labels):
        return 'wip' not in labels and 'do not merge' not in labels

    @staticmethod
    def _get_mergeable_prs(repo):
        pulls = repo.get_pulls(state='open', sort='created', base='master')
        mergeable_prs = []

        for pull in pulls:
            if pull.mergeable_state in ['clean', 'behind']:
                labels = [label.name for label in pull.get_labels()]

                if PrMerge._not_wip_and_merge_allowed(labels):
                    print(f'PR-{pull.number} is potentially mergeable adding to the list.')
                    mergeable_prs.append(pull)
                else:
                    print(f'PR-{pull.number} has the label(s) {labels}. Skipping...')

            elif pull.mergeable_state == 'blocked':
                print(f'PR-{pull.number} has not been approved. Skipping...')
            elif pull.mergeable_state == 'dirty':
                print(f'PR-{pull.number} has conflicts that need to be manually resolved. Skipping...')
            else:
                print(f'PR-{pull.number} has merge status {pull.mergeable_state} which is not handled. Skipping...')

        return mergeable_prs

    @staticmethod
    def _merge_pr(mergeable_pr):
        print(f'Merging PR {mergeable_pr.number} {mergeable_pr.title}')
        merge_status = mergeable_pr.merge(merge_method='merge')

        print(f'PR merge status Merged: {merge_status.merged} \n'
              f'                Message: {merge_status.message} \n'
              f'                SHA: {merge_status.sha}')
