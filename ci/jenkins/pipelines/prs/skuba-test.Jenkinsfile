/**
 * This pipeline verifies on a Github PR:
 *   - skuba unit and integration tests
 *   - Basic skuba deployment, bootstrapping, and adding nodes to a cluster
 */
def platformNames = ['OpenStack']
def platformTests = [:]
def platformTest(platformName) {
    return {
        environment {
            OPENRC = credentials('ecp-openrc')
            PLATFORM = platformName.toLowerCase()
            CLUSTERNAME = "${PLATFORM}-cluster"
        }

        sh(script: 'make -f skuba/ci/Makefile create_environment', label: 'Create Environment')
        archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfstate")
        archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfvars.json")
        sshagent(credentials: ['shared-ssh-key']) {
            dir('skuba') {
                sh(script: "SKUBA_BIN_PATH=\"${WORKSPACE}/go/bin/skuba\" GINKGO_BIN_PATH=\"${WORKSPACE}/skuba/ginkgo\" IP_FROM_TF_STATE=TRUE make test-e2e", label: "End-to-end tests")
            }
        }

        post {
            always {
                sh(script: 'make -f skuba/ci/Makefile gather_logs', label: 'Gather Logs')
                zip(archive: true, dir: "testrunner_${PLATFORM}_logs", zipFile: "testrunner_${PLATFORM}_logs.zip")
            }
            cleanup {
                sh(script: 'make -f skuba/ci/Makefile destroy_environment', label: 'Destroy Environment')
            }
        }
    }
}

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        GITHUB_TOKEN = credentials('github-token')
        STACK_NAME = "${JOB_NAME}-${BUILD_NUMBER}"
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

        stage('Get Worker Info') { steps {
            sh(script: 'make -f skuba/ci/Makefile info', label: 'Info')
        } }

        stage('PR Checks') { steps {
            sh(script: 'make -f skuba/ci/Makefile pr_checks', label: 'PR Checks')
        } }

        stage('Environment Setup') { steps {
            dir('skuba') {
                sh(script: 'make build-ginkgo', label: 'Build Ginkgo Binary')
            }
        } }

        stage('Run skuba unit tests') { steps {
            dir("skuba") {
              sh(script: 'make test-unit', label: 'make test-unit')
            }
        } }

        stage('Build Skuba') { steps {
            sh(script: 'make -f skuba/ci/Makefile create_skuba', label: 'Create Skuba')
        } }

        stage('Platform Tests') {
            steps {
                script {

                    if(sh(script: "skuba/${PR_MANAGER} filter-pr --filename ci/infra/vmware", returnStdout: true, label: "Filtering PR") =~ "contains changes") {
                        platformNames << 'VMware'
                    }
                    platformNames.each {platformName ->
                        platformTests.put(platformName, platformTest(platformName))
                    }
                    parallel(platformTests)
                }
            }
        }
    }

    post {
        always {
            sh(script: 'make -f skuba/ci/Makefile cleanup', label: 'Cleanup')
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
