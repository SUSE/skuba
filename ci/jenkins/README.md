# CaaSP Jenkins Job Builder

1. Copy `ci/jenkins/jenkins_jobs.ini.template` to `ci/jenkins/jenkins_jobs.ini` 
2. Fill in with the needed values.

#### Run with:
`make test_jenkins_jobs` to test that the jobs are created correctly 
will output xml to ci/jenkins/jobs

`make update_jenkins_jobs` to update the jobs in jenkins