/**
 * This pipeline verifies for the basic caaspctl deployment, bootstrapping, and add nodes for cluster on a regular branch
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack'
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile pre_deployment', label: 'Pre Deployment')
        } }

        stage('Cluster Deployment') { steps {
            sh(script: 'make -f caaspctl/ci/Makefile deploy', label: 'Deploy')
            archiveArtifacts("caaspctl/ci/infra/${PLATFORM}/terraform.tfstate")
        } }

        stage('Run end-to-end tests') { steps {
            dir("caaspctl") {
              sh(script: 'IP_FROM_TF_STATE=TRUE PLATFORM=openstack make test-e2e', label: 'End-to-end tests')
        }} }

   }
   post {
       always {
           sh(script: 'make --keep-going -f caaspctl/ci/Makefile post_run', label: 'Post Run')
       }
       cleanup {
           dir("${WORKSPACE}") {
               deleteDir()
           }
       }
    }
}
