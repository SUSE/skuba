/**
 * This pipeline perform basic checks on Pull-requests. (go vet) etc
 */

pipeline {
    agent { node { label 'caasp-team-private-integration' } }

    environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PR_CONTEXT = 'jenkins/skuba-code-lint'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
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

        stage('Running make lint') { steps {
            sh(script: 'make lint', label: 'make lint')
        } }
        stage('Checking status of git tree') { steps {
            sh(script: 'test -z "$(git status --porcelain go.mod go.sum vendor/)" || { echo "there are uncommitted changes. This should never happen"; exit 1; }', label: 'git tree status')
        } }

        // TODO: Add here golint later on

    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        unstable {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        failure {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
