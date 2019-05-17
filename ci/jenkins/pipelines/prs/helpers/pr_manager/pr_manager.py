#!/usr/bin/env Python3

import argparse
import configparser
import os
import sys

from github import Github

from pr_merge import PrMerge

GITHUB_TOKEN = os.getenv('GITHUB_TOKEN')
JENKINS_CONFIG = configparser.ConfigParser()
GITHUB_REPO = 'SUSE/skuba'

if GITHUB_TOKEN is None:
    print('Env var GITHUB_TOKEN missing please set it')
    sys.exit(1)
else:
    # Make sure to only grab the token
    GITHUB_TOKEN = GITHUB_TOKEN.split(':')[-1]


def _read_config(config_path):
    if config_path is None:
        print('No config provided please set JENKINS_CONFIG or use the --config arg')
    JENKINS_CONFIG.read(config_path)


def merge_prs(args):
    if args.config:
        _read_config(args.config)
    else:
        _read_config(os.getenv('JENKINS_CONFIG'))

    g = Github(GITHUB_TOKEN)
    repo = g.get_repo(GITHUB_REPO)

    merger = PrMerge(JENKINS_CONFIG, repo)
    merger.merge_prs()


def parse_args():
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers()

    # Parse merge-prs command
    merge_parser = subparsers.add_parser('merge-prs', help='Look for and merge a Pull Request')
    merge_parser.add_argument('--config', help='Path to ini config for interacting with Jenkins')
    merge_parser.set_defaults(func=merge_prs)

    parsed_args = parser.parse_args()

    return parsed_args


if __name__ == '__main__':
    args = parse_args()
    args.func(args)
