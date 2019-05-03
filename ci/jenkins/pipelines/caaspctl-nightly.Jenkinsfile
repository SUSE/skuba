/**
 * This pipeline verifies for the basic caaspctl deployment, bootstrapping, and add nodes for cluster on a regular branch
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENRC = "/srv/ci/ecp.openrc"
        GITHUB_TOKEN = readFile("/srv/ci/github-token").trim()
        PARAMS = "stack-type=openstack-terraform"
   }

   stages {
        stage('Git Clone') { steps {
            sh "rm ${WORKSPACE}/* -rf"
            sh "git clone https://${GITHUB_TOKEN}@github.com/SUSE/caaspctl"
        } }

        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile pre_deployment', label: 'Pre Deployment')
        } }

        stage('Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile deploy', label: 'Deploy')
        } }

        stage('Bootstrap Cluster') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile bootstrap', label: 'Bootstrap')
        } }

        stage('Add Nodes to Cluster') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile add_nodes', label: 'Add Nodes')
        } }
   }
   post {
       always {
           sh(script: 'make -f caaspctl/ci/Makefile post_run', label: 'Post Run')
           cleanWs()
       }
    }
}
