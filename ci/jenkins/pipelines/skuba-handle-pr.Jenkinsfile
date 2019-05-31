/**
 * This pipeline merges the PRs
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
       GITHUB_TOKEN = credentials('github-token')
       JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
       PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
       REQUESTS_CA_BUNDLE = '/var/lib/ca-certificates/ca-bundle.pem'
   }

   stages {
        stage('Examining PRs to merge') { steps {
            sh(script: "${PR_MANAGER} merge-prs --config ${JENKINS_JOB_CONFIG}", label: 'Checking ready PRs')
        } }

   }
   post {
       cleanup {
           dir("${WORKSPACE}") {
               deleteDir()
           }
       }
    }
}
