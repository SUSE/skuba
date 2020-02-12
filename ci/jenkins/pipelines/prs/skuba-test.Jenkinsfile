/**
 * This pipeline verifies on a Github PR:
 *   - skuba unit and e2e tests
 *   - Basic skuba deployment, bootstrapping, and adding nodes to a cluster
 */

pipeline {
    agent { node { label 'caasp-team-private-integration' } }

    environment {
        SKUBA_BINPATH = '/home/jenkins/go/bin/skuba'
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack'
        TERRAFORM_STACK_NAME = "${BUILD_NUMBER}-${JOB_NAME.replaceAll("/","-")}".take(70)
        PR_CONTEXT = 'jenkins/skuba-test'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        REQUESTS_CA_BUNDLE = '/var/lib/ca-certificates/ca-bundle.pem'
        PIP_VERBOSE = 'true'
    }


    stages {
        stage('Collaborator Check') { steps { script {
            if (env.BRANCH_NAME.startsWith('PR')) {
                def membersResponse = httpRequest(
                    url: "https://api.github.com/repos/SUSE/skuba/collaborators/${CHANGE_AUTHOR}",
                    authentication: 'github-token',
                    validResponseCodes: "204:404")

                if (membersResponse.status == 204) {
                    echo "Test execution for collaborator ${CHANGE_AUTHOR} allowed"

                } else {
                    def allowExecution = false

                    try {
                        timeout(time: 5, unit: 'MINUTES') {
                            allowExecution = input(id: 'userInput', message: "Change author is not a SUSE member: ${CHANGE_AUTHOR}", parameters: [
                                booleanParam(name: 'allowExecution', defaultValue: false, description: 'Run tests anyway?')
                            ])
                        }
                    } catch(err) {
                        def user = err.getCauses()[0].getUser()
                        if('SYSTEM' == user.toString()) {
                            echo "Timeout while waiting for input"
                        } else {
                            allowExecution = false
                            echo "Unhandled error:\n${err}"
                        }
                    }
                    

                    if (!allowExecution) {
                        echo "Test execution for unknown user (${CHANGE_AUTHOR}) disallowed"
                        error(message: "Test execution for unknown user (${CHANGE_AUTHOR}) disallowed")
                        return;
                    }
                }
            }
        } } }

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

        stage('Run skuba unit tests') { steps {
            dir("skuba") {
              sh(script: 'make test-unit', label: 'make test-unit')
            }
        } }

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'make -f skuba/ci/Makefile pr_checks', label: 'PR Checks')
            sh(script: "pushd skuba; make -f Makefile install; popd", label: 'Build Skuba')
        } }

        stage('Cluster Provisioning') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile create_environment', label: 'Provision')
            }
        }

       stage('Run Pre Bootstrap Tests') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile test_pre_bootstrap', label: 'Test Pre Bootstrap')
           }
       }

        stage('Cluster Bootstrap') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile bootstrap', label: 'Bootstrap')
            }
        }

       stage('Run Post Bootstrap Tests') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile test_post_bootstrap', label: 'Test Post Bootstrap')
           }
       }

        stage('Join Nodes') {
            steps {
                sh(script: 'make -f skuba/ci/Makefile join_nodes', label: 'Join Nodes')
            }
        }


        stage('Run e2e tests') { steps {
            sh(script: "make -f skuba/ci/Makefile test_e2e", label: "test_e2e")
        } }

    }
    post {
        always {
            archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfstate", allowEmptyArchive: true)
            archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfvars.json", allowEmptyArchive: true)
            archiveArtifacts(artifacts: 'testrunner.log', allowEmptyArchive: true)
            archiveArtifacts(artifacts: 'skuba/ci/infra/testrunner/*.xml', allowEmptyArchive: true)
            sh(script: "make --keep-going -f skuba/ci/Makefile gather_logs", label: 'Gather Logs')
            archiveArtifacts(artifacts: 'platform_logs/**/*', allowEmptyArchive: true)
            junit('skuba/ci/infra/testrunner/*.xml')
        }
        cleanup {
            sh(script: "make --keep-going -f skuba/ci/Makefile cleanup", label: 'Cleanup')
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
            sh(script: "rm -f ${SKUBA_BINPATH}; ", label: 'Remove built skuba')
        }
        unstable {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        failure {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
