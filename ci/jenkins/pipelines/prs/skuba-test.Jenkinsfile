/**
 * This pipeline verifies on a Github PR:
 *   - skuba unit and integration tests
 *   - Basic skuba deployment, bootstrapping, and adding nodes to a cluster
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack'
        TERRAFORM_STACK_NAME = "${JOB_NAME.replaceAll("/","-")}-${BUILD_NUMBER}"
        PR_CONTEXT = 'jenkins/skuba-test'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        REQUESTS_CA_BUNDLE = '/var/lib/ca-certificates/ca-bundle.pem'
    }

    stages {
        stage('Setting GitHub in-progress status') { steps {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'pending'", label: "Sending pending status")
        } }

        stage('Git Clone') { steps {
            deleteDir()
            checkout([$class: 'GitSCM',
                      branches: [[name: "*/${BRANCH_NAME}"], [name: '*/master']],
                      doGenerateSubmoduleConfigurations: false,
                      extensions: [[$class: 'LocalBranch'],
                                   [$class: 'WipeWorkspace'],
                                   [$class: 'RelativeTargetDirectory', relativeTargetDir: 'skuba']],
                      submoduleCfg: [],
                      userRemoteConfigs: [[refspec: '+refs/pull/*/head:refs/remotes/origin/PR-*',
                                           credentialsId: 'github-token',
                                           url: 'https://github.com/SUSE/skuba']]])

            dir("${WORKSPACE}/skuba") {
                sh(script: "git checkout ${BRANCH_NAME}", label: "Checkout PR Branch")
            }
        }}

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'make -f skuba/ci/Makefile pr_checks', label: 'PR Checks')
        } }


        stage('Run All Skuba Upgrade Tests In Parallel') {
            parallel {
                stage('Run Skuba Upgrade Plan All fine Test') {
                    steps {
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-plan-all-fine-${BUILD_NUMBER}; make -f skuba/ci/Makefile test_upgrade_plan_all_fine", label: 'Skuba Upgrade Plan All fine')
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-plan-all-fine-${BUILD_NUMBER}; make --keep-going -f skuba/ci/Makefile post_run", label: 'Post Run')
                    }
                }

                stage('Run Skuba Upgrade Plan from previous Test') {
                    steps {
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-plan-from-previous-${BUILD_NUMBER}; make -f skuba/ci/Makefile test_upgrade_plan_from_previous", label: 'Skuba Upgrade Plan From Previous')
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-plan-from-previous-${BUILD_NUMBER}; make --keep-going -f skuba/ci/Makefile post_run", label: 'Post Run')
                    }
                }

                stage('Run Skuba Upgrade Apply All fine Test') {
                    steps {
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-apply-all-fine-${BUILD_NUMBER}; make -f skuba/ci/Makefile test_upgrade_apply_all_fine", label: 'Skuba Upgrade Apply All fine')
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-apply-all-fine-${BUILD_NUMBER}; make --keep-going -f skuba/ci/Makefile post_run", label: 'Post Run')
                    }
                }

                stage('Run Skuba Upgrade Apply from previous Test') {
                    steps {
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-apply-from-previous-${BUILD_NUMBER}; make -f skuba/ci/Makefile test_upgrade_apply_from_previous", label: 'Skuba Upgrade Apply From Previous')
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-apply-from-previous-${BUILD_NUMBER}; make --keep-going -f skuba/ci/Makefile post_run", label: 'Post Run')
                    }
                }

                stage('Run Skuba Upgrade Apply With User Lock') {
                    steps {
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-apply-user-lock-${BUILD_NUMBER}; make -f skuba/ci/Makefile test_upgrade_apply_user_lock", label: 'Skuba Upgrade Apply With User Lock')
                        sh(script: "export TERRAFORM_STACK_NAME=${JOB_NAME}-apply-user-lock-${BUILD_NUMBER}; make --keep-going -f skuba/ci/Makefile post_run", label: 'Post Run')
                    }
                }
            }
        }

    }
    post {
        always {
            sh(script: 'make --keep-going -f skuba/ci/Makefile post_run', label: 'Post Run')
            zip(archive: true, dir: 'testrunner_logs', zipFile: 'testrunner_logs.zip')
        }
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
