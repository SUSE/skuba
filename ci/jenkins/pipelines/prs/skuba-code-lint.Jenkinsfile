/**
 * This pipeline perform basic checks on Pull-requests. (go vet) etc
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PR_CONTEXT = 'jenkins/skuba-code-lint'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
    }

    stages {
        stage('Setting GitHub in-progress status') { steps {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'pending'", label: "Sending pending status")
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
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
