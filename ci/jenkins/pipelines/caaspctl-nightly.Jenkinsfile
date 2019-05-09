/**
 * This pipeline verifies for the basic caaspctl deployment, bootstrapping, and add nodes for cluster on a regular branch
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PARAMS = "stack-type=openstack-terraform"
   }

   stages {
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
       }
       cleanup {
           dir("${WORKSPACE}") {
               deleteDir()
           }
       }
    }
}
