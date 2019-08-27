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

    def check_pr_details(self, pr_number):
        """
        Performs the following checks for every commits in a given PR
        - Checks to verify that SUSE employees are using their @suse email for their commits.
        - Checks that the commit body is not empty
        :param pr_number: Number of the PR to look up
        :return:
        """
        pr = self.repo.get_pull(pr_number)
        email_pattern = re.compile(r'^.*@suse\.(com|cz|de)$')

        for commit in pr.get_commits():
            sha = commit.sha
            author = commit.author
            title = message = commit.commit.message
            # Not sure why we need to use the nested commit for the email
            email = commit.commit.author.email
            user_id = f'{author.login}({email})'
            body = ''

            # This could be probably smarter but commit contains something like the following
            # message="$commit_title\n\n$long_commit_message" and as such maybe we can split it and
            # check for the following limits: title max 50 chars, body max 72 chars per line and at
            # least as long as the commit title to avoid commit message bodies full of whitespaces
            try:
                title, body = message.split('\n\n', 1)
            except ValueError:
                print('No commit body was detected')

            print(f'Checking commit "{sha}: {title}"')

            if not email_pattern.fullmatch(email):
                print(f'Checking if {user_id} is part of the SUSE organization...')

                if self.org.has_in_members(commit.author):
                    print(f'{user_id} is part of SUSE organization but a SUSE e-mail address was not used for commit: {sha}')
                    sys.exit(1)

            title = title.split('(bsc#') # Title may contain (bsc#XXXXXXXX) references so we need to exclude these
            if len(title[0].rstrip()) > 50:
                print('Commit message title should be less than 50 characters (excluding the bsc# reference)')
                sys.exit(1)

            # No body detected. Nothing else to do here.
            if not body:
                continue

            if len(body) < len(title):
                print('Commit message body is too short')
                sys.exit(1)

            for body_line in body.splitlines():
                if len(body_line) > 72:
                    print('Each line in the commit body should be less than 72 characters')
                    sys.exit(1)

        print(f'PR-{pr_number} commits verified.')
