import sys
import re


class PrChecks:
    def __init__(self, org, repo):
        self.org = org
        self.repo = repo

    def check_pr_from_fork(self, pr_number):
        """
        Checks if the given PR is from a fork if not exit with an error code
        All PRs should come from forks not the main repo.

        :param pr_number: Number of the PR to look up
        :return:
        """
        pr = self.repo.get_pull(pr_number)

        if pr.head.repo.full_name == self.repo.full_name:
            print(f'PR-{pr_number} is coming from a branch in the target repository. This is not allowed!')
            print('Please send your PR from a forked repository instead.')
            sys.exit(1)

        print(f'PR-{pr_number} is from a fork.')

    def check_employee_emails(self, pr_number):
        """
        Checks to verify that SUSE employees are using their @suse email for their commits.
        :param pr_number: Number of the PR to look up
        :return:
        """
        pr = self.repo.get_pull(pr_number)
        email_pattern = re.compile(r'^.*@suse\.(com|cz|de)$')

        for commit in pr.get_commits():
            sha = commit.sha
            author = commit.author
            # Not sure why we need to use the nested commit for the email
            email = commit.commit.author.email
            user_id = f'{author.login}({email})'

            if email_pattern.fullmatch(email):
                print(f'Commit {sha} is from SUSE employee {user_id}. Moving on...')
                continue

            print(f'Checking if {user_id} is part of the SUSE organization...')

            if self.org.has_in_members(commit.author):
                print(f'{user_id} is part of SUSE organization but a SUSE e-mail address was not used for commit: {sha}')
                sys.exit(1)

        print(f'PR-{pr_number} commit email(s) verified.')
