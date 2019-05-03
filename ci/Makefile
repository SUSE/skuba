.DEFAULT_GOAL := all
MKFILE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
TEST_RUNNER := ${MKFILE_DIR}/infra/testrunner/testrunner
JENKINS_JOBS_DIR = ${MKFILE_DIR}/jenkins
JENKINS_JOB_CONFIG ?= ${JENKINS_JOBS_DIR}/jenkins_jobs.ini

PARAMS ?= stack-type=openstack-terraform

.PHONY: all
all: deploy bootstrap

# Steps
.PHONY: bootstrap_environment
bootstrap_environment:
	${TEST_RUNNER} stage=bootstrap_environment ${PARAMS}

.PHONY: configure_environment
configure_environment:
	${TEST_RUNNER} stage=configure_environment ${PARAMS}

.PHONY: create_environment
create_environment:
	${TEST_RUNNER} stage=create_environment ${PARAMS}

.PHONY: final_cleanup
final_cleanup:
	${TEST_RUNNER} stage=final_cleanup ${PARAMS}

.PHONY: gather_logs
gather_logs:
	${TEST_RUNNER} stage=gather_logs ${PARAMS}

.PHONY: git_rebase_check
git_rebase_check:
	${TEST_RUNNER} stage=git_rebase branch-name=${BRANCH_NAME} ${PARAMS}

.PHONY: github_collaborator_check
github_collaborator_check:
	${TEST_RUNNER} stage=github_collaborator_check ${PARAMS}

.PHONY: grow_environment
grow_environment:
	${TEST_RUNNER} stage=grow_environment ${PARAMS}

.PHONY: info
info:
	${TEST_RUNNER} stage=info ${PARAMS}

.PHONY: initial_cleanup
initial_cleanup:
	${TEST_RUNNER} stage=initial_cleanup ${PARAMS}

.PHONY: test_jenkins_jobs
test_jenkins_jobs:
	jenkins-jobs --ignore-cache test ${JENKINS_JOBS_DIR}/:${JENKINS_JOBS_DIR}/templates/ -o ${JENKINS_JOBS_DIR}/jobs/ --config-xml

.PHONY: update_jenkins_jobs
update_jenkins_jobs:
	jenkins-jobs  --ignore-cache --conf ${JENKINS_JOB_CONFIG} update ${JENKINS_JOBS_DIR}/:${JENKINS_JOBS_DIR}/templates/

# Stages
.PHONY: pre_deployment
pre_deployment: info initial_cleanup

.PHONY: pr_checks
pr_checks: github_collaborator_check git_rebase_check

.PHONY: deploy
deploy: create_environment configure_environment

.PHONY: bootstrap
bootstrap: bootstrap_environment

.PHONY: add_nodes
add_nodes: grow_environment

.PHONY: post_run
post_run: gather_logs final_cleanup
