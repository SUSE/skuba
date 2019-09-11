/**
 * This pipeline runs any e2e test parametrized
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   parameters {
        string(name: 'E2E_MAKE_TARGET_NAME', defaultValue: 'all', description: 'The make target to run (only e2e related)')
   }

   environment {
        SKUBA_BINPATH = "/home/jenkins/go/bin/skuba"
        OPENSTACK_OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        VMWARE_ENV_FILE = credentials('vmware-env')
        TERRAFORM_STACK_NAME = "${BUILD_NUMBER}-${JOB_NAME.replaceAll("/","-")}".take(80)
   }

   stages {
        stage('Getting Ready For Cluster Deployment') { steps {
            sh(script: "make -f skuba/ci/Makefile pre_deployment", label: 'Pre Deployment')
            sh(script: "pushd skuba; make -f Makefile install; popd", label: 'Build Skuba')
        } }

        stage('Run Skuba e2e Test') {
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
            sh(script: "rm -f ${SKUBA_BINPATH}; ", label: 'Remove built skuba')
       }
    }
}
