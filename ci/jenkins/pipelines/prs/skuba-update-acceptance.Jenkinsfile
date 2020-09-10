// this pipeline runs os acceptance tests for skuba-update

pipeline {
    agent { node { label 'caasp-team-private-integration' } }
    environment {
        GITHUB_TOKEN = credentials('github-token')
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
        PR_CONTEXT = 'jenkins/skuba-update-acceptance'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        FILTER_SUBDIRECTORY = 'skuba-update'
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
            sh(script: "${PR_MANAGER} update-pr-status ${PR_CONTEXT} 'pending'", label: "Sending pending status")
        } }

        stage('skuba-update SUSE OS Tests') {
            when {
                expression {
                    sh(script: "${PR_MANAGER} filter-pr --filename ${FILTER_SUBDIRECTORY}", returnStdout: true, label: "Filtering PR") =~ "contains changes"
                }
            }
            steps {
                sh(script: "make -f skuba-update/test/os/suse/Makefile test", label: 'skuba-update SUSE OS Tests')
            }
        }
    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
                sh(script: 'sudo rm -rf skuba-update/build skuba-update/skuba_update.egg-info', label: 'Remove python artifacts created by root')
                sh(script: 'sudo rm -rf skuba-update/test/os/suse/artifacts', label: 'Remove test artifacts created by root')
                deleteDir()
            }
        }
        unstable {
            sh(script: "${PR_MANAGER} update-pr-status ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        failure {
            sh(script: "${PR_MANAGER} update-pr-status ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "${PR_MANAGER} update-pr-status ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
