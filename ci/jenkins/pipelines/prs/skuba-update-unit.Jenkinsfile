// this pipeline runs unit tests for skuba-update

pipeline {
    agent { node { label 'caasp-team-private-integration' } }
    environment {
        GITHUB_TOKEN = credentials('github-token')
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
        PR_CONTEXT = 'jenkins/skuba-update-unit'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        FILTER_SUBDIRECTORY = 'skuba-update'
        DOCKERIZED_UNIT_TESTS = '1'
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

        stage('skuba-update SUSE Unit Tests') {
            when {
                expression {
                    sh(script: "${PR_MANAGER} filter-pr --filename ${FILTER_SUBDIRECTORY}", returnStdout: true, label: "Filtering PR") =~ "contains changes"
                }
            }
            steps {
                dir("skuba-update") {
                    sh(script: "make test", label: 'skuba-update Unit Tests')
                }
            }
        }
    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
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
