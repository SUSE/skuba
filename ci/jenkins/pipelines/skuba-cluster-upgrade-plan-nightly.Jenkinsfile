/**
 * This pipeline verifies skuba cluster upgrade plan
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        TERRAFORM_STACK_NAME = "${JOB_NAME}-${BUILD_NUMBER}"
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
        } }

        stage('Run Skuba Cluster Upgrade Plan Test') { steps {
            sh(script: 'make -f skuba/ci/Makefile test_cluster_upgrade_plan_all_fine', label: 'Skuba Cluster Upgrade Plan')
        } }
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
