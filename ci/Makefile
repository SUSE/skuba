.DEFAULT_GOAL := all
MKFILE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
TEST_RUNNER := ${MKFILE_DIR}/infra/testrunner/testrunner
JENKINS_JOBS_DIR = ${MKFILE_DIR}/jenkins
JENKINS_JOB_CONFIG ?= ${JENKINS_JOBS_DIR}/jenkins_jobs.ini
PLATFORM ?= openstack

.PHONY: all
all: deploy bootstrap

# Steps
.PHONY: add_master_nodes
add_master_nodes:
	${TEST_RUNNER} --add-nodes --master 2 --vars vars/${PLATFORM}.yaml

.PHONY: add_worker_nodes
add_worker_nodes:
	${TEST_RUNNER} --add-nodes --worker 1 --vars vars/${PLATFORM}.yaml

.PHONY: bootstrap_environment
bootstrap_environment:
	${TEST_RUNNER} --bootstrap --vars vars/${PLATFORM}.yaml

.PHONY: create_skuba
create_skuba:
	${TEST_RUNNER} --create-skuba --vars vars/${PLATFORM}.yaml

.PHONY: create_environment
create_environment:
	${TEST_RUNNER} --terraform-apply --vars vars/${PLATFORM}.yaml

.PHONY: cleanup
cleanup:
	${TEST_RUNNER} --cleanup --vars vars/${PLATFORM}.yaml

.PHONY: gather_logs
gather_logs:
	${TEST_RUNNER} --log --vars vars/${PLATFORM}.yaml

.PHONY: git_rebase_check
git_rebase_check:
	${TEST_RUNNER} --git-rebase --vars vars/${PLATFORM}.yaml

.PHONY: info
info:
	${TEST_RUNNER} --info --vars vars/${PLATFORM}.yaml

.PHONY: test_jenkins_jobs
test_jenkins_jobs:
	jenkins-jobs --ignore-cache test ${JENKINS_JOBS_DIR}/:${JENKINS_JOBS_DIR}/templates/ -o ${JENKINS_JOBS_DIR}/jobs/ --config-xml

.PHONY: update_jenkins_jobs
update_jenkins_jobs:
	jenkins-jobs  --ignore-cache --conf ${JENKINS_JOB_CONFIG} update ${JENKINS_JOBS_DIR}/:${JENKINS_JOBS_DIR}/templates/

# Stages
.PHONY: pre_deployment
pre_deployment: info

.PHONY: pr_checks
pr_checks: git_rebase_check

.PHONY: deploy
deploy: create_environment create_skuba

.PHONY: bootstrap
bootstrap: bootstrap_environment

.PHONY: add_nodes
add_nodes: add_master_nodes add_worker_nodes

.PHONY: post_run
post_run: gather_logs cleanup
