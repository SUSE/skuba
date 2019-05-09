# Jenkins pipelines caasp

## prs:

This directory contain all pipelines used for GitHub PULL-Request validation.

### PR Pipeline Functions:

**setBuildStatus(String context, String description, String state)**

Due to the way our CI is currently set up this function is necessary 
to set the status separately for each pipeline.

Uses the GitHub API https://developer.github.com/v3/repos/statuses/ 


## others:

the pipelines outside the prs directory are intented to perform regular operations for master branch or regular jobs not on GitHub PRs
