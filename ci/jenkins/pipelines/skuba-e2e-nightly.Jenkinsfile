/**
 * This pipeline runs any e2e test parametrized
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   properties([
       parameters([
           string(name: 'E2E_MAKE_TARGET_NAME', defaultValue: 'all', description: 'The make target to run (only e2e related)')
       ])
   ])

   environment {
        SKUBA_BINPATH = "/home/jenkins/go/bin/skuba"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
        TERRAFORM_STACK_NAME = "${JOB_NAME}-${E2E_MAKE_TARGET_NAME}-${BUILD_NUMBER}"
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: "make -f skuba/ci/Makefile pre_deployment", label: 'Pre Deployment')
            sh(script: "pushd skuba; make -f Makefile install; popd", label: 'Build Skuba')
        } }

        stage("Run Skuba ${E2E_MAKE_TARGET_NAME} Test") {
            steps {
                sh(script: "make -f skuba/ci/Makefile ${E2E_MAKE_TARGET_NAME}", label: "${E2E_MAKE_TARGET_NAME}")
                sh(script: "make --keep-going -f skuba/ci/Makefile cleanup", label: 'Cleanup')
            }
        }
   }

   post {
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
       }
    }
}
