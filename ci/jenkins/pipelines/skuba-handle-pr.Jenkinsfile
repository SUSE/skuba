/**
 * This pipeline merges the PRs
 */

pipeline {
   agent { node { label 'caasp-team-private' } }

   environment {
        GITHUB_TOKEN = credentials('github-token')
   }

   stages {
        stage('Examining PRs to merge') { steps {
            sh(script: 'ci/jenkins/pipelines/prs/helpers/handle-prs.sh', label: 'Checking ready PRs')
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
