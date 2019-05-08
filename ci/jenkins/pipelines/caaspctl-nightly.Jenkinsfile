/**
 * This pipeline verifies for the basic caaspctl deployment, bootstrapping, and add nodes for cluster on a regular branch
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENRC = "/srv/ci/ecp.openrc"
        GITHUB_TOKEN = readFile("/srv/ci/github-token").trim()
        PARAMS = "stack-type=openstack-terraform no-collab-check"
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh "ci/infra/testrunner/testrunner stage=info ${PARAMS}"
            sh "ci/infra/testrunner/testrunner stage=github_collaborator_check ${PARAMS}"
            sh "ci/infra/testrunner/testrunner stage=initial_cleanup ${PARAMS}"
        } }

        stage('Cluster Deployment') { steps {
            sh "ci/infra/testrunner/testrunner stage=create_environment ${PARAMS}"
            sh "ci/infra/testrunner/testrunner stage=configure_environment ${PARAMS}"
        } }

        stage('Bootstrap Cluster') { steps {
            sh "ci/infra/testrunner/testrunner stage=bootstrap_environment ${PARAMS}"
        } }

        stage('Add Nodes in Cluster') { steps {
            sh "ci/infra/testrunner/testrunner stage=grow_environment ${PARAMS}"
        } }
   }
   post {
        always {
            sh "ci/infra/testrunner/testrunner stage=gather_logs ${PARAMS}"
            sh "ci/infra/testrunner/testrunner stage=final_cleanup ${PARAMS}"
            cleanWs()
        }
    }
}
