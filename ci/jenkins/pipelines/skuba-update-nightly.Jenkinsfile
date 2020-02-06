/**
 * This pipeline verifies for the basic skuba-update os tests
 */

pipeline {
   agent { node { label 'caasp-team-private-integration' } }

   environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        ENV_FILE = credentials('vmware-env')
   }

   stages {
        stage('skuba-update SUSE OS Tests') { steps {
            sh(script: "make -f skuba/skuba-update/test/os/suse/Makefile test", label: 'skuba-update SUSE OS Tests')
        } }
   }
   post {
       cleanup {
            dir("${WORKSPACE}") {
                sh(script: 'sudo rm -rf skuba/skuba-update/build skuba/skuba-update/skuba_update.egg-info', label: 'Remove python artifacts created by root')
                sh(script: 'sudo rm -rf skuba/skuba-update/test/os/suse/artifacts', label: 'Remove test artifacts created by root')
                deleteDir()
            }
        }
    }
}
