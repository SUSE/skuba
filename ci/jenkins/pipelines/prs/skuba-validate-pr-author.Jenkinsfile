/**
 * This pipeline performs various check for PR authors
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        GITHUB_TOKEN = credentials('github-token')
        PR_CONTEXT = 'jenkins/skuba-validate-pr-author'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
    }

    stages {
        stage('Collaborator Check') { steps {
            sh(script: "${PR_MANAGER} check-pr --collab-check", label: "Checking if collaborator")
        }}

        stage('Setting GitHub in-progress status') { steps {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'pending'", label: "Sending pending status")
        } }

        stage('Validating PR author') { steps {
            sh(script: "${PR_MANAGER} check-pr --is-fork --check-pr-details", label: 'checking valid PR author')
        } }

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
