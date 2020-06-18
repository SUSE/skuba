/*
 * This pipeline verifies for the basic skuba-update os tests
 */

// type of worker required by the job 
def worker_type = 'integration'

node('caasp-team-private-integration') {
    stage('select worker') {
        if (env.BRANCH != 'master') {
            if (env.BRANCH.startsWith('experimental') || env.BRANCH.startsWith('maintenance')) {
                worker_type = env.BRANCH
            }
        }
    }
}

pipeline {
   agent { node { label "caasp-team-private-${worker_type} && e2e" } }

   environment {
        OPENRC = credentials('ecp-openrc')
        GITHUB_TOKEN = credentials('github-token')
        ENV_FILE = credentials('vmware-env')
   }

   stages {
        stage('skuba-update SUSE OS Tests') { steps {
            sh(script: "make -f skuba-update/test/os/suse/Makefile test", label: 'skuba-update SUSE OS Tests')
        } }
   }
   post {
       cleanup {
            dir("${WORKSPACE}") {
                sh(script: 'sudo rm -rf skuba-update/build skuba-update/skuba_update.egg-info', label: 'Remove python artifacts created by root')
                sh(script: 'sudo rm -rf skuba-update/test/os/suse/artifacts', label: 'Remove test artifacts created by root')
                deleteDir()
            }
        }
    }
}
