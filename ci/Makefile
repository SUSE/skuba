.DEFAULT_GOAL := all
MKFILE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
TEST_RUNNER := ${MKFILE_DIR}/infra/testrunner/testrunner
TASKS := ${MKFILE_DIR}/tasks
REBASE := rebase.py
JENKINS_JOBS_DIR = ${MKFILE_DIR}/jenkins
JENKINS_JOB_CONFIG ?= ${JENKINS_JOBS_DIR}/jenkins_jobs.ini
PLATFORM ?= openstack

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

# Stages
.PHONY: pre_deployment
pre_deployment: info

.PHONY: pr_checks
pr_checks: git_rebase_check

.PHONY: deploy
deploy: create_environment

.PHONY: bootstrap
bootstrap: bootstrap_environment

.PHONY: post_run
post_run: gather_logs cleanup

# e2e Tests
.PHONY: test_upgrade_plan_all_fine
test_upgrade_plan_all_fine:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_skuba_upgrade.py --test test_upgrade_plan_all_fine

.PHONY: test_upgrade_plan_from_previous
test_upgrade_plan_from_previous:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_skuba_upgrade.py --test test_upgrade_plan_from_previous

.PHONY: test_upgrade_apply_all_fine
test_upgrade_apply_all_fine:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_skuba_upgrade.py --test test_upgrade_apply_all_fine

.PHONY: test_upgrade_apply_from_previous
test_upgrade_apply_from_previous:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_skuba_upgrade.py --test test_upgrade_apply_from_previous

.PHONY: test_upgrade_apply_user_lock
test_upgrade_apply_user_lock:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_skuba_upgrade.py --test test_upgrade_apply_user_lock

.PHONY: test_cilium
test_cilium:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_cilium.py

.PHONY: test_deployments
test_deployments:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_deployments.py

.PHONY: test_dockercaps
test_dockercaps:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --suite test_dockercaps.py

.PHONY: test_integration
test_integration:
	${TEST_RUNNER} -v vars.yaml -p ${PLATFORM} test --verbose --junit --filter "not disruptive and not flaky" --skip-setup deployed
