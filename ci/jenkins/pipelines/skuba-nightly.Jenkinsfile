/**
 * This pipeline verifies for the basic skuba deployment, bootstrapping, and add nodes for cluster on a regular branch
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        ENV_FILE = credentials('vmware-env')
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
        } }

        stage('Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile deploy', label: 'Deploy')
            archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfstate")
            archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfvars")
        } }

       stage('Run end-to-end tests') { steps {
           dir("skuba") {
             sh(script: 'make build-ginkgo', label: 'build ginkgo binary')
             sh(script: "make setup-ssh && SKUBA_BIN_PATH=\"${WORKSPACE}/go/bin/skuba\" GINKGO_BIN_PATH=\"${WORKSPACE}/skuba/ginkgo\" IP_FROM_TF_STATE=TRUE PLATFORM=openstack make test-e2e", label: "End-to-end tests")
       } } }
 
   }
   post {
       always {
           sh(script: 'make --keep-going -f skuba/ci/Makefile post_run', label: 'Post Run')
       }
       cleanup {
           dir("${WORKSPACE}") {
               deleteDir()
           }
       }
    }
}
