// this pipeline runs unit tests for skuba-update

pipeline {
    agent { node { label 'caasp-team-private' } }
    environment {
        GITHUB_TOKEN = credentials('github-token')
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
        PR_CONTEXT = 'jenkins/skuba-update-unit'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        SUBDIRECTORY = 'skuba-update'
        DOCKERIZED_UNIT_TESTS = '1'
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

        stage('skuba-update SUSE Unit Tests') {
            when {
                expression {
                    stdout = sh(script: 'skuba/ci/jenkins/pipelines/prs/helpers/pr-filter.sh', label: 'checking if PR contains skuba-update changes')
                    return (stdout =~ "contains changes")
                }
            }
            steps {
                dir("skuba/skuba-update") {
                    sh(script: "make test", label: 'skuba-update Unit Tests')
                }
            }
        }
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
