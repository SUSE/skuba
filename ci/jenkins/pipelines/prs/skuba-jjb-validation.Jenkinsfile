// this pipeline update all jenkins pipelines via jenkins job builder plugin

pipeline {
    agent { node { label 'caasp-team-private' } }
    environment {
        GITHUB_TOKEN = credentials('github-token')
        JENKINS_JOB_CONFIG = credentials('jenkins-job-config')
        REQUESTS_CA_BUNDLE = "/var/lib/ca-certificates/ca-bundle.pem"
        PR_CONTEXT = 'jenkins/skuba-jjb-validation'
        PR_MANAGER = 'ci/jenkins/pipelines/prs/helpers/pr-manager'
        TERRAFORM_STACK_NAME = "${BUILD_NUMBER}-${JOB_NAME.replaceAll("/","-")}".take(70)
    }
    stages {

        stage('Collaborator Check') { steps {
            sh(script: "${PR_MANAGER} check-pr --collab-check", label: "Checking if collaborator")
        }}

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

        stage('Info') { steps {
            sh(script: "make -f skuba/ci/Makefile info", label: 'Info')
        } }

        stage('Setup Environment') { steps {
            sh(script: 'python3 -m venv venv', label: 'Setup Python Virtualenv')
            sh(script: 'venv/bin/pip install jenkins-job-builder==2.10.0', label: 'Install Dependencies')
        } }

        stage('Test Jobs') { steps {
            sh(script: """
                   source ${WORKSPACE}/venv/bin/activate
                   make -f skuba/ci/Makefile test_jenkins_jobs
                """, label: 'Test Jenkins Jobs')
            zip archive: true, dir: 'skuba/ci/jenkins/jobs', zipFile: 'jenkins_jobs.zip'
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
