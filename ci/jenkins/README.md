# CI Jenkins 
 
This directory contains all of the scripts and pipelines for our CI in Jenkins.

All jobs are automatically updated using Jenkins Job Builder 
 
## Jenkins Job Builder

1. Copy `ci/jenkins/jenkins_jobs.ini.template` to `ci/jenkins/jenkins_jobs.ini` 
2. Fill in with the needed values.

### Run with:
`make test_jenkins_jobs` to test that the jobs are created correctly 
will output xml to ci/jenkins/jobs

`make update_jenkins_jobs` to update the jobs in jenkins

## Pipelines

### PRs:

This directory contains all the pipelines used for GitHub Pull Request validation.

#### PRs Helpers:

Scripts found under here are meant to be ran by Jenkins **NOT** users.

##### pr-manager

This script is for handling PR interactions between Jenkins and GitHub.

It uses the same config as Jenkins Job Builder to interact with Jenkins



### Others:

The pipelines outside the prs directory are intended to perform regular operations for master branch or regular jobs not on GitHub PRs
