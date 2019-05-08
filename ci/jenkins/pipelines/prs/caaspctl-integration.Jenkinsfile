/**
 * This pipeline verifies basic caaspctl deployment, bootstrapping, and adding nodes to a cluster on GitHub Pr
 */

pipeline {
    agent { node { label 'caasp-team-private' } }

    environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PARAMS = "stack-type=openstack-terraform no-collab-check"
    }

    stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: "ci/infra/testrunner/testrunner stage=info ${PARAMS}", label: "Info")
            sh(script: "ci/infra/testrunner/testrunner stage=github_collaborator_check ${PARAMS}", label: "GitHub Collaborator Check")
            sh(script: "ci/infra/testrunner/testrunner stage=git_rebase branch-name=${env.BRANCH_NAME} ${PARAMS}", label: "Git Rebase Check")
            sh(script: "ci/infra/testrunner/testrunner stage=initial_cleanup ${PARAMS}", label: "Initial Cleanup")
        } }

        stage('Cluster Deployment') { steps {
            sh(script: "ci/infra/testrunner/testrunner stage=create_environment ${PARAMS}", label: "Create Environment")
            sh(script: "ci/infra/testrunner/testrunner stage=configure_environment ${PARAMS}", label: "Configure Environment")
        } }

        stage('Bootstrap Cluster') { steps {
            sh(script: "ci/infra/testrunner/testrunner stage=bootstrap_environment ${PARAMS}", label: "Bootstrap Environment")
        } }

        stage('Add Nodes in Cluster') { steps {
            sh(script: "ci/infra/testrunner/testrunner stage=grow_environment ${PARAMS}", label: "Grow Environment")
        } }
    }
    post {
        always {
            sh(script: "ci/infra/testrunner/testrunner stage=gather_logs ${PARAMS}", label: "Gather Logs")
            sh(script: "ci/infra/testrunner/testrunner stage=final_cleanup ${PARAMS}", label: "Final Cleanup")
        }
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
    }
}
