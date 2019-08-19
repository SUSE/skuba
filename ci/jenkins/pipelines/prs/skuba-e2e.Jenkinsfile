/**
 * This pipeline runs any e2e test parametrized
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        SKUBA_BINPATH = "/home/jenkins/go/bin/skuba"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
        TERRAFORM_STACK_NAME = "${JOB_NAME.replaceAll("/","-")}-${BUILD_NUMBER}"
        PR_CONTEXT = 'jenkins/skuba-test'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        REQUESTS_CA_BUNDLE = '/var/lib/ca-certificates/ca-bundle.pem'
    }

    stages {
        stage('Jenkins Worker Info') {
            steps {
                sh(script: "make -f skuba/ci/Makefile info", label: 'Info')
            }
        }

        stage('Setting GitHub in-progress status') {
            steps {
                sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'pending'", label: "Sending pending status")
            }
        }

        stage('Git Clone') {
            steps {
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
            }
        }

        stage('Getting Ready For Cluster Deployment') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile pr_checks', label: 'PR Checks')
                sh(script: "pushd skuba; make -f Makefile install; popd", label: 'Build Skuba')
            }
        }

        stage('Run Skuba e2e Test') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile test_pr_e2e', label: 'test_pr_e2e')
                archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfstate")
                archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfvars.json")
            }
        }
    }

    post {
        always {
            archiveArtifacts("testrunner_logs/**/*")
            sh(script: "make --keep-going -f skuba/ci/Makefile cleanup", label: 'Cleanup')
        }
        cleanup {
            dir("${WORKSPACE}@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}@script") {
                deleteDir()
            }
            dir("${WORKSPACE}@script@tmp") {
                deleteDir()
            }
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
