/**
 * This pipeline perform basic checks on Pull-requests. (go vet) etc
 */

void setBuildStatus(String context, String description, String state) {
    def body = "{\"state\": \"${state}\", " +
               "\"target_url\": \"${BUILD_URL}/display/redirect\", " +
               "\"description\": \"${description}\", " +
               "\"context\": \"${context}\"}"
    def headers = '-H "Content-Type: application/json" -H "Accept: application/vnd.github.v3+json"'
    def url = "https://${GITHUB_TOKEN}@api.github.com/repos/SUSE/caaspctl/statuses/${GIT_COMMIT}"

    sh(script: "curl -X POST ${headers} ${url} -d '${body}'", label: "Sending commit status")
}

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PARAMS = "stack-type=openstack-terraform no-collab-check"
    }

    stages {
        stage('Setting GitHub in-progress status') { steps {
            setBuildStatus('jenkins/caaspctl-code-lint', 'in-progress', 'pending')
        } }

        stage('Running go vet') { steps {
            sh(script: 'make vet', label: 'Go Vet')
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
            setBuildStatus('jenkins/caaspctl-code-lint', 'failed', 'failure')
        }
        success {
            setBuildStatus('jenkins/caaspctl-code-lint', 'success', 'success')
        }
    }
}
