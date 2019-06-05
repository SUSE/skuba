// this pipeline runs os integration tests for skuba-update

pipeline {
    agent { node { label 'caasp-team-private' } }
    environment {
        GITHUB_TOKEN = credentials('github-token')
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
        PR_CONTEXT = 'jenkins/skuba-update-integration'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
    }
    stages {
        stage('Setting GitHub in-progress status') { steps {
            sh(script: "${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'pending'", label: "Sending pending status")
        } }

        stage('Git Clone') { steps {
            deleteDir()
            checkout([$class: 'GitSCM',
                      branches: [[name: "*/${BRANCH_NAME}"]],
                      doGenerateSubmoduleConfigurations: false,
                      extensions: [[$class: 'LocalBranch'],
                                   [$class: 'WipeWorkspace'],
                                   [$class: 'RelativeTargetDirectory', relativeTargetDir: 'skuba']],
                      submoduleCfg: [],
                      userRemoteConfigs: [[refspec: '+refs/pull/*/head:refs/remotes/origin/PR-*',
                                           credentialsId: 'github-token',
                                           url: 'https://github.com/SUSE/skuba']]])
        }}

        stage('skuba-update SUSE OS Tests') { steps {
            sh(script: "make -f skuba/skuba-update/test/os/suse/Makefile test", label: 'skuba-update SUSE OS Tests')
        } }
    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
                deleteDir()
            }
        }
        failure {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'failure'", label: "Sending failure status")
        }
        success {
            sh(script: "skuba/${PR_MANAGER} update-pr-status ${GIT_COMMIT} ${PR_CONTEXT} 'success'", label: "Sending success status")
        }
    }
}
