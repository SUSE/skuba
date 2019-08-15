/**
 * This pipeline verifies for the basic skuba deployment, bootstrapping, and add nodes for cluster on a regular branch
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        SKUBA_BIN_PATH = "/home/jenkins/go/bin/skuba"
        GINKGO_BIN_PATH = "${WORKSPACE}/skuba/ginkgo"
        IP_FROM_TF_STATE = "TRUE"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        TERRAFORM_STACK_NAME = "${JOB_NAME.replaceAll("/","-")}-${BUILD_NUMBER}"
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: 'cd skuba; make install; cd ../', label: 'Install skuba')
        } }

        stage('Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile deploy', label: 'Deploy')
            archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfstate")
            archiveArtifacts("skuba/ci/infra/${PLATFORM}/terraform.tfvars.json")
        } }

       stage('Run end-to-end tests') { steps {
           dir("skuba") {
             sh(script: 'make build-ginkgo', label: 'Build ginkgo binary')
             sh(script: "make setup-ssh", label: "Setup SSH")
             sh(script: "make test-e2e", label: "End-to-end tests")
       } } }
 
   }
   post {
       always {
           sh(script: 'make --keep-going -f skuba/ci/Makefile post_run', label: 'Post Run')
           zip(archive: true, dir: 'testrunner_logs', zipFile: 'testrunner_logs.zip')
       }
       cleanup {
           dir("${WORKSPACE}") {
               deleteDir()
           }
       }
    }
}
