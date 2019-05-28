/**
 * This pipeline performs various check for PR authors
 */

void setBuildStatus(String context, String description, String state) {
    def body = "{\"state\": \"${state}\", " +
               "\"target_url\": \"${BUILD_URL}/display/redirect\", " +
               "\"description\": \"${description}\", " +
               "\"context\": \"${context}\"}"
    def headers = '-H "Content-Type: application/json" -H "Accept: application/vnd.github.v3+json"'
    def url = "https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/skuba/statuses/${GIT_COMMIT}"

    sh(script: "curl -X POST ${headers} ${url} -d '${body}'", label: "Sending commit status")
}

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        GITHUB_TOKEN = credentials('github-token')
    }

    stages {
        stage('Setting GitHub in-progress status') { steps {
            setBuildStatus('jenkins/skuba-validate-pr-author', 'in-progress', 'pending')
        } }

        stage('Validating PR author') { steps {
            sh(script: 'ci/jenkins/pipelines/prs/helpers/check-valid-author.sh', label: 'checking valid PR author')
        } }

    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            setBuildStatus('jenkins/skuba-validate-pr-author', 'failed', 'failure')
        }
        success {
            setBuildStatus('jenkins/skuba-validate-pr-author', 'success', 'success')
        }
    }
}
