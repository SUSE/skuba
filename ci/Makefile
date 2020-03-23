.DEFAULT_GOAL := all
MKFILE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
TEST_RUNNER := ${MKFILE_DIR}/infra/testrunner/testrunner
TASKS := ${MKFILE_DIR}/tasks
REBASE := rebase.py
JENKINS_JOBS_DIR = ${MKFILE_DIR}/jenkins
JENKINS_JOB_CONFIG ?= ${JENKINS_JOBS_DIR}/jenkins_jobs.ini
PLATFORM ?= openstack
ifdef SUITE
SUITE_CMD=--suite ${SUITE}.py
endif
ifdef TEST
ifndef SUITE
$(error SUITE is required for selecting a test )
endif
TEST_CMD=--test ${TEST}
endif

.PHONY: all
all: deploy bootstrap

# Steps

.PHONY: bootstrap_environment
bootstrap_environment:
	${TEST_RUNNER} --platform ${PLATFORM} bootstrap

.PHONY: create_environment
create_environment:
	${TEST_RUNNER} --platform ${PLATFORM} provision

.PHONY: cleanup
cleanup:
	${TEST_RUNNER} --platform ${PLATFORM} cleanup

.PHONY: gather_logs
gather_logs:
	${TEST_RUNNER} --platform ${PLATFORM} get_logs

.PHONY: git_rebase_check
git_rebase_check:
	${TASKS}/${REBASE}

.PHONY: info
info:
	${TEST_RUNNER} --platform ${PLATFORM} info

.PHONY: join_nodes
join_nodes:
	${TEST_RUNNER} --platform ${PLATFORM} join-nodes

.PHONY: test_jenkins_jobs
test_jenkins_jobs:
	jenkins-jobs --ignore-cache test ${JENKINS_JOBS_DIR}/:${JENKINS_JOBS_DIR}/templates/ -o ${JENKINS_JOBS_DIR}/jobs/ --config-xml

.PHONY: update_jenkins_jobs
update_jenkins_jobs:
	jenkins-jobs  --ignore-cache --conf ${JENKINS_JOB_CONFIG} update ${JENKINS_JOBS_DIR}/:${JENKINS_JOBS_DIR}/templates/

.PHONY: update_pkgs
update_pkgs:
	sudo zypper update --allow-vendor-change -y kubernetes-client terraform terraform-provider-*

# Stages
.PHONY: pre_deployment
pre_deployment: info update_pkgs

.PHONY: pr_checks
pr_checks: git_rebase_check

.PHONY: deploy
deploy: create_environment

.PHONY: bootstrap
bootstrap: bootstrap_environment

.PHONY: post_run
post_run: gather_logs cleanup

# e2e Tests
.PHONY: test
test:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose ${SUITE_CMD} ${TEST_CMD}

.PHONY: test_pr
test_pr:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --junit pr --filter "pr" --skip-setup deployed

.PHONE: inhibit_kured
inhibit_kured:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} inhibit_kured

