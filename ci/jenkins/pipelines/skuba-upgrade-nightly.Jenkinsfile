/**
 * This pipeline verifies skuba node/cluster upgrade plan/apply
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: 'make -f skuba/ci/Makefile pre_deployment', label: 'Pre Deployment')
        } }

        parallel {
            stage('Run Skuba Upgrade All fine Test') {
                environment {
                    TERRAFORM_STACK_NAME = "${JOB_NAME}-all-fine-${BUILD_NUMBER}"
                }

                steps {
                    sh(script: 'make -f skuba/ci/Makefile test_upgrade_all_fine', label: 'Skuba Upgrade All fine')
                }
            }

            stage('Run Skuba Upgrade from previous Test') {
                environment {
                    TERRAFORM_STACK_NAME = "${JOB_NAME}-from-previous-${BUILD_NUMBER}"
                }

                steps {
                    sh(script: 'make -f skuba/ci/Makefile test_upgrade_from_previous', label: 'Skuba Upgrade from previous')
                }
            }
        }
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
