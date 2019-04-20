/**
 * This pipeline verifies for the basic caaspctl deployment, bootstrapping, and add nodes for cluster.
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENRC = "/srv/ci/ecp.openrc"
        GITHUB_TOKEN = readFile("/srv/ci/github-token").trim()
        PARAMS = "stack-type=openstack-terraform no-collab-check"
   }

   stages {
        stage('Git Clone') { steps {
            sh "rm ${WORKSPACE}/* -rf"
            sh "git clone https://${GITHUB_TOKEN}@github.com/SUSE/caaspctl"
        } }

        stage('Getting Ready For Cluster Deployment') { steps {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=info ${PARAMS}"
            sh "caaspctl/ci/infra/testrunner/testrunner stage=github_collaborator_check ${PARAMS}"
            sh "caaspctl/ci/infra/testrunner/testrunner stage=initial_cleanup ${PARAMS}"
        } }

        stage('Cluster Deployment') { steps {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=create_environment ${PARAMS}"
            sh "caaspctl/ci/infra/testrunner/testrunner stage=configure_environment ${PARAMS}"
        } }

        stage('Bootstrap Cluster') { steps {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=bootstrap_environment ${PARAMS}"
        } }

        stage('Add Nodes in Cluster') { steps {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=grow_environment ${PARAMS}"
        } }
   }
   post {
        always {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=gather_logs ${PARAMS}"
            sh "caaspctl/ci/infra/testrunner/testrunner stage=final_cleanup ${PARAMS}"
            cleanWs()
        }
    }
}
