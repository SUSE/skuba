/**
 * This pipeline verifies for the basic skuba deployment, bootstrapping, and add nodes for cluster on a regular branch
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        SKUBA_BINPATH = "/home/jenkins/go/bin/skuba"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        TERRAFORM_STACK_NAME = "${JOB_NAME.replaceAll("/","-")}-${BUILD_NUMBER}"
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
            sh(script: "pushd skuba; make -f Makefile install; popd", label: 'Build Skuba')
        } }

       stage('Cluster Provisioning') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile create_environment', label: 'Provision')
               archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfstate", allowEmptyArchive: true)
               archiveArtifacts(artifacts: "skuba/ci/infra/${PLATFORM}/terraform.tfvars.json", allowEmptyArchive: true)
           }
       }

       stage('Run Pre Bootstrap Tests') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile test_pre_bootstrap', label: 'Test Pre Bootstrap')
           }
       }

       stage('Cluster Bootstrap') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile bootstrap', label: 'Bootstrap')
           }
       }

       stage('Run Post Bootstrap Tests') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile test_post_bootstrap', label: 'Test Post Bootstrap')
           }
       }

       stage('Join Nodes') {
           steps {
               sh(script: 'make -f skuba/ci/Makefile join_nodes', label: 'Join Nodes')
           }
       }

       stage('Run Integration tests') { steps {
           sh(script: "make -f skuba/ci/Makefile test_integration", label: "test_integration")
       } }

   }
   post {
        always {
            junit('skuba/ci/infra/testrunner/*.xml')
            sh(script: "make --keep-going -f skuba/ci/Makefile post_run", label: 'Post Run')
            archiveArtifacts(artifacts: 'testrunner_logs/**/*', allowEmptyArchive: true)
            archiveArtifacts(artifacts: 'testrunner.log', allowEmptyArchive: true)
        }
        cleanup {
            dir("${WORKSPACE}@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}@script") {
                deleteDir()
            }
            dir("${WORKSPACE}@script@tmp") {
                deleteDir()
            }
            dir("${WORKSPACE}") {
                deleteDir()
            }
        sh(script: "rm -f ${SKUBA_BINPATH}; ", label: 'Remove built skuba')
       }
    }
}
