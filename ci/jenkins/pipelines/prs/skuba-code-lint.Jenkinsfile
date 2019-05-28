/**
 * This pipeline perform basic checks on Pull-requests. (go vet) etc
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
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PARAMS = "stack-type=openstack-terraform"
    }

    stages {
        stage('Setting GitHub in-progress status') { steps {
            setBuildStatus('jenkins/skuba-code-lint', 'in-progress', 'pending')
        } }

        stage('Running make lint') { steps {
            sh(script: 'make lint', label: 'make lint')
        } }

        // TODO: Add here golint later on

    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            setBuildStatus('jenkins/skuba-code-lint', 'failed', 'failure')
        }
        success {
            setBuildStatus('jenkins/skuba-code-lint', 'success', 'success')
        }
    }
}
