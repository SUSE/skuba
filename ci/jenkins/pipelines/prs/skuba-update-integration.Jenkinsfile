// this pipeline runs os integration tests for skuba-update

pipeline {
    agent { node { label 'caasp-team-private' } }
    environment {
        GITHUB_TOKEN = credentials('github-token')
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
        PR_CONTEXT = 'jenkins/skuba-update-integration'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        FILTER_SUBDIRECTORY = 'skuba-update'
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

        stage('skuba-update SUSE OS Tests') {
            when {
                expression {
                    stdout = sh(script: 'skuba/ci/jenkins/pipelines/prs/helpers/pr-filter.sh', label: 'checking if PR contains skuba-update changes')
                    return (stdout =~ "contains changes")
                }
            }
            steps {
                sh(script: "make -f skuba/skuba-update/test/os/suse/Makefile test", label: 'skuba-update SUSE OS Tests')
            }
        }
    }
    post {
        cleanup {
            dir("${WORKSPACE}") {
                sh(script: 'sudo rm -rf skuba/skuba-update/build skuba/skuba-update/skuba_update.egg-info', label: 'Remove python artifacts created by root')
                sh(script: 'sudo rm -rf skuba/skuba-update/test/os/suse/artifacts', label: 'Remove test artifacts created by root')
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
